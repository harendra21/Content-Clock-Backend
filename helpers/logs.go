package helpers

func Logging(logType, message string) {
	app := CreateApp()

	if logType == "" {
		app.Logger().Info(message)
		return
	}
	if logType == "error" {
		app.Logger().Error(message)
		return
	}
	if logType == "info" {
		app.Logger().Info(message)
		return
	}
	if logType == "debug" {
		app.Logger().Debug(message)
		return
	}
	if logType == "warn" {
		app.Logger().Warn(message)
		return
	}

}
