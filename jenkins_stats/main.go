package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/bndr/gojenkins"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Build struct {
	gorm.Model
	BuildId  string
	JobName  string
	Status   string
	Duration int64
}

func save(db gorm.DB, build gojenkins.Build, jobName string) {
	var found Build
	db.First(&found, "build_id=? and job_name=?", build.Raw.ID, jobName)
	if found.ID == 0 {
		db.Create(
			&Build{
				BuildId:  build.Raw.ID,
				JobName:  jobName,
				Status:   build.GetResult(),
				Duration: int64(build.GetDuration()),
			},
		)
	} else {
		found.Duration = int64(build.GetDuration())
		found.Status = build.GetResult()
		db.Save(found)
	}
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

	if len(os.Args) == 1 {
		fmt.Println("Supply job name as first and only argument")
	} else {
		jobName := os.Args[1]
		jobs, err := jenkins.GetAllBuildIds(ctx, jobName)
		if err != nil {
			panic(fmt.Sprintln("cannot retrieve builds for job", jobName, ":", err))
		}

		count := 0
		for _, job := range jobs {
			count++
			buildId := job.Number
			build, err := jenkins.GetBuild(ctx, jobName, buildId)

			if err != nil {
				panic(fmt.Sprintln("cannot retrieve build", buildId, " for job", jobName, ":", err))
			}

			fmt.Printf("Build [%s/%d]: %s [%s]\n", jobName, buildId, build.GetResult(), time.UnixMilli(build.Raw.Timestamp).Format(time.RFC1123Z))
			save(*db, *build, jobName)
		}
		fmt.Println("found", count, "builds for", jobName)
	}
}
