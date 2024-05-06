package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bndr/gojenkins"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Build struct {
	gorm.Model
	Id     uint16
	Status string
}

func main() {
	db, err := gorm.Open(sqlite.Open("/tmp/jenkins.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&Build{})

	ctx := context.Background()
	jenkins := gojenkins.CreateJenkins(nil, os.Getenv("JENKINS_URL"), os.Getenv("JENKINS_USER"), os.Getenv("JENKINS_PASSWORD"))
	_, err = jenkins.Init(ctx)
	if err != nil {
		panic(fmt.Sprintln("failed to connect to jenkins:", err))
	}

	db.Create(&Build{Id: 1, Status: "success"})

	var build Build
	db.First(&build, 1)
	db.First(&build, "id = ?", 1)
}
