package main

import (
	"content-clock/controllers"
	"content-clock/helpers"
	"log"

	"github.com/joho/godotenv"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

var app *pocketbase.PocketBase

func main() {
	// app := pocketbase.New()

	app = helpers.CreateApp()

	godotenv.Load()

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// serves static files from the provided public dir (if exists)
		// se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))
		controllers.SetupInstagramRoutes(se, app)
		controllers.SetupFacebookRoutes(se, app)
		controllers.SetupLinkedinRoutes(se, app)
		controllers.SetupTwitterRoutes(se, app)
		controllers.SetupMastodonRoutes(se, app)
		controllers.SetupPinterestRoutes(se, app)
		controllers.SetupAiRoutes(se, app)
		controllers.SetupRedditRoutes(se, app)
		controllers.SetupThreadsRoutes(se, app)
		return se.Next()
	})

	app.Cron().MustAdd("Publish Scheduled Posts", "* * * * *", func() {
		controllers.GetScheduledPosts(app)
	})
	app.Cron().MustAdd("Fetch Analytics (3 Hrs)", "0 */3 * * *", func() {
		controllers.FetchPostsAnalytics(app)
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
