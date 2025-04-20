// Package main это точка входа в приложение
package main

import "internshipPVZ/cmd/di_container"

func main() {
	appModule := di_container.AppModule{}
	appModule.Invoke()
}
