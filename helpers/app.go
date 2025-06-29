package helpers

import "github.com/pocketbase/pocketbase"

func CreateApp() *pocketbase.PocketBase {
	app := pocketbase.NewWithConfig(pocketbase.Config{
		HideStartBanner: false,
	})

	return app
}
