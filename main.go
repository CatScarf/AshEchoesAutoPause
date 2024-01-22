package main

import (
	"fmt"
	"image"
	"log"
	"runtime"
	"time"
)

type Status int

const (
	Error Status = iota
	OutGame
	InGame
	Combat
	Pause
)

func (s Status) String() string {
	return [...]string{"Error", "OutGame", "InGame", "Combat", "Pause"}[s]
}

type DetectResult struct {
	status       Status
	validMonitor int
	founded      map[string]float64
	pauseX       float32
	pauseY       float32
}

// 检测游戏状态
func detectStatus(crops map[string]Crop, patterns *map[string]image.Image) DetectResult {

	result := DetectResult{}

	window, err := FindWindowByTitle(`白荆回廊\[[0-9.]+\]`)
	if err != nil {
		result.status = OutGame
		return result
	}

	windowRect := GetClientAreaRect(window.Handle)
	pauseX, pauseY := crops["pause"].GetCenter(windowRect)
	result.pauseX = pauseX
	result.pauseY = pauseY

	screenshots, err := GetScreenShots()
	if err != nil {
		log.Panicf("GetScreenShots failed: %v", err)
		result.status = Error
		return result
	}

	maxFounded := make(map[string]float64)

	maxSimilarity := -1.0
	validMonitor := -1

	for i, screenshot := range screenshots {
		cutAndMatch := func(targetName string, patternName string) float64 {
			target := crops[targetName].CutImage(&screenshot, windowRect)
			pattern := (*patterns)[patternName]
			return MatchImage(&target, pattern)
		}

		founded := make(map[string]float64)
		founded["pause"] = cutAndMatch("pause", "pause")
		founded["resume"] = cutAndMatch("pause", "resume")
		founded["substitude"] = cutAndMatch("substitude", "substitude")
		founded["settings"] = cutAndMatch("settings", "settings")

		sumSimilarity := 0.0
		for _, similarity := range founded {
			sumSimilarity += similarity
		}

		if sumSimilarity > maxSimilarity {
			maxSimilarity = sumSimilarity
			maxFounded = founded
			validMonitor = i
		}

	}

	result.status = InGame
	result.validMonitor = validMonitor
	result.founded = maxFounded

	threshold := 0.7

	if maxFounded["substitude"] < threshold && maxFounded["settings"] < threshold {
		return result
	}

	if maxFounded["pause"] > threshold {
		result.status = Combat
	} else if maxFounded["resume"] > threshold {
		result.status = Pause
	}

	return result
}

func main() {
	isSuperUser := true

	if runtime.GOOS == "windows" {
		isSuperUser = true
		if !isAdministrator() {
			fmt.Println("请右键以管理员身份运行...")
			fmt.Scanf("%s")
			isSuperUser = false
			return
		}
	}

	crops := make(map[string]Crop)
	crops["pause"] = Crop{0.933, 0.034, 0.05, 1.6, false}
	crops["substitude"] = Crop{0.848, 0.101, 0.05, 2.5, false}
	crops["settings"] = Crop{0.025, 0.945, 0.025, 1, true}

	patterns := make(map[string]image.Image)
	// patterns["pause"] = LoadImage("pause.png")
	// patterns["resume"] = LoadImage("resume.png")
	// patterns["substitude"] = LoadImage("substitude.png")
	// patterns["settings"] = LoadImage("settings.png")
	patterns["pause"] = LoadBase64Image(pauseBase64)
	patterns["resume"] = LoadBase64Image(resumeBase64)
	patterns["substitude"] = LoadBase64Image(substitudeBase64)
	patterns["settings"] = LoadBase64Image(settingsBase64)

	var fps float64 = 10
	lastStatus := Error

	for {
		startTime := time.Now()

		result := detectStatus(crops, &patterns)

		fmt.Printf("[%7s] %s\n", result.status, BeautifulMap(result.founded))

		if result.status == Combat && lastStatus == InGame {
			ClickAndBack(int(result.pauseX), int(result.pauseY), 10)
			superUserHint := ""
			if !isSuperUser {
				superUserHint = "（未以管理员身份运行，操作失败）"
			} else {
				superUserHint = ""
			}
			fmt.Printf("Click %d, %d%s\n", int(result.pauseX), int(result.pauseY), superUserHint)
		}
		lastStatus = result.status

		elapsed := float64(time.Since(startTime).Milliseconds())
		if elapsed <= 1000.0/fps {
			time.Sleep(time.Duration(1000/fps-elapsed) * time.Millisecond)
		} else {
			fmt.Printf("Out of time: %f\n", 000.0/fps-elapsed)
		}
	}

}
