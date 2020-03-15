package common

import "log"

func ErrorHandler(message string, err error) {
	if err != nil {
		log.Println("Failed: " + message + "," + err.Error())
	}
}
