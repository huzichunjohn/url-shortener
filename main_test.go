package main

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/kataras/iris/httptest"
)

func TestURLShortener(t *testing.T) {
	f, err := ioutil.TempFile("", "shortener")
	if err != nil {
		t.Fatalf("creating temp file for database failed: %v", err)
	}

	db := NewDB(f.Name())
	app := newApp(db)

	e := httptest.New(t, app)
	originalURL := "https://google.com"

	e.POST("/shorten").
		WithFormField("url", originalURL).Expect().
		Status(httptest.StatusOK).Body().Contains("<pre><a target='_new' href=")
	
	keys := db.GetByValue(originalURL)
	if got := len(keys); got != 1 {
		t.Fatalf("expected to have 1 key but saved %d short urls", got)
	}

	e.GET("/u/" + keys[0]).Expect().
		Status(httptest.StatusTemporaryRedirect).Header("Location").Equal(originalURL)

	e.POST("/shorten").
		WithFormField("url", originalURL).Expect().
		Status(httptest.StatusOK).Body().Contains("<pre><a target='_new' href=")

	keys2 := db.GetByValue(originalURL)
	if got := len(keys2); got != 1 {
		t.Fatalf("expected to have 1 keys even if we save the same original url but saved %d short urls", got)
	}

	if keys[0] != keys2[0] {
		t.Fatalf("expected keys to be equal if the original url is the same, but got %s = %s ", keys[0],  keys2[0])
	}

	e.POST("/clear_cache").Expect().Status(httptest.StatusOK)
	if got := db.Len(); got != 0 {
		t.Fatalf("expected database to have 0 registered objects after /clear_cache but has %d", got)
	}

	db.Close()
	time.Sleep(1*time.Second)
	if err := f.Close(); err != nil {
		t.Fatalf("unable to close the file: %s", f.Name())
	}

	if err := os.Remove(f.Name()); err != nil {
		t.Fatalf("unable to remove the file from %s", f.Name())
	}

	time.Sleep(1*time.Second)
}