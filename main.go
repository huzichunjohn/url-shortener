package main

import (
	"html/template"

	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

func main() {
	db := NewDB("shortener.db")
	app := newApp(db)
	iris.RegisterOnInterrupt(db.Close)
	app.Run(iris.Addr(":8080"))
}

func newApp(db *DB) *iris.Application {
	app := iris.Default()
	factory := NewFactory(DefaultGenerator, db)
	tmpl := iris.HTML("./templates", ".html").Reload(true)
	tmpl.AddFunc("IsPositive", func(n int) bool {
		if n > 0 {
			return true
		}
		return false
	})
	app.RegisterView(tmpl)
	app.StaticWeb("/static", "./resources")
	indexHandler := func(ctx context.Context) {
		ctx.ViewData("URL_COUNT", db.Len())
		ctx.View("index.html")
	}
	app.Get("/", indexHandler)
	execShortURL := func(ctx context.Context, key string) {
		if key == "" {
			ctx.StatusCode(iris.StatusBadRequest)
			return
		}

		value := db.Get(key)
		if value == "" {
			ctx.StatusCode(iris.StatusNotFound)
			ctx.Writef("Short URL for key: '%s' not found", key)
			return
		}

		ctx.Redirect(value, iris.StatusTemporaryRedirect)
	}
	app.Get("/u/{shortkey}", func(ctx context.Context) {
		execShortURL(ctx, ctx.Params().Get("shortkey"))
	})
	app.Post("/shorten", func(ctx context.Context) {
		formValue := ctx.FormValue("url")
		if formValue == "" {
			ctx.ViewData("FORM_RESULT", "You need to enter a URL")
			ctx.StatusCode(iris.StatusLengthRequired)
		} else {
			key, err := factory.Gen(formValue)
			if err != nil {
				ctx.ViewData("FORM_RESULT", "Invalid URL")
				ctx.StatusCode(iris.StatusBadRequest)
			} else {
				if err = db.Set(key, formValue); err != nil {
					ctx.ViewData("FORM_RESULT", "Internal error while saving the URL")
					app.Logger().Infof("while saving URL: " + err.Error())
					ctx.StatusCode(iris.StatusInternalServerError)
				} else {
					ctx.StatusCode(iris.StatusOK)
					shortenURL := "http://" + app.ConfigurationReadOnly().GetVHost() + "/u/" + key
					ctx.ViewData("FORM_RESULT", 
						template.HTML("<pre><a target='_new' href='"+shortenURL+"'>"+shortenURL+"</a></pre>"))
				}
			}
		}
		
		indexHandler(ctx)
	})
	app.Post("/clear_cache", func(ctx context.Context) {
		db.Clear()
		ctx.Redirect("/")
	})

	return app
}