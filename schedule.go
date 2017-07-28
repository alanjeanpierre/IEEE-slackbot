package main

import(
	"errors"
	"strings"
	"strconv"
	"os"
	"fmt"
	"io/ioutil"
	"bufio"
	
	
)

// readSchedule accepts a cleaned string from caller
// starting from M 5 3 6 3 T 5 2 W 5 etc
func readSchedule(input string, free bool, rootloc string, id string) (err error) {

	var spec int = 0;
	//var d2i map[string]int{'m' : 0, 't': 1, 'w': 2, 'r': 3, 'f': 4}
	
	var schedule [11][5]int
	if free {
		spec = 1
	} else {
		for i, row := range schedule {
			for j := range row {
				schedule [i][j] = 1
			}
		}
	}
	
	
	f := func(c rune) bool {
		return 	c == 'm' ||
				c == 't' ||
				c == 'w' ||
				c == 'r' ||
				c == 'f'
	}
	
	days := strings.FieldsFunc(input, f)
	if len(days) != 5 {
		err = errors.New("invalid fields")
		return
	}
	
	for day, timestr := range days {
		times := strings.Fields(timestr)
		for _, time := range times {
		
			time, errr := strconv.Atoi(time)
			if time == 0 {
				continue
			}
			if time > 17 || time < 8 || errr != nil {
				err = errors.New("Invalid times")
				return
			}
			schedule[time-8][day] = spec
		}
	}
	
	return saveSchedule(schedule, rootloc, id)
}

func saveSchedule(schedule [11][5]int, rootloc string, id string) (err error) {
	
	file, err := os.OpenFile(rootloc + id + ".sched", os.O_WRONLY | os.O_CREATE, 0664)
	if err != nil {
		return
	}
	defer file.Close()
	
	for _, row := range schedule {
		for _, column := range row {
			fmt.Fprintf(file, "%d ", column)
		}
		fmt.Fprintf(file, "\n")
	}
	
	return
}

func bestTime(rootloc string) (day string, time int, num int) {
	files, err := ioutil.ReadDir(rootloc)
	if err != nil {
		return
	}

	var availables [11][5]int
	var max int
	var maxj int
	var maxi int
	i2day := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"}
	
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sched") {
			sched := loadSchedule(rootloc + file.Name())
			for i, row := range sched {
				for j, column := range row {
					availables[j][i] += column
					if (availables[j][i] > max) {
						maxj = j
						maxi = i
						max = availables[j][i]
					}
				}
			}
		}
	}
	
	return i2day[maxi], maxj + 8, max
	
	
}

func loadSchedule(path string) (schedule [5][11]int) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0664)
	if err != nil {
		return
	}
	
	scanner := bufio.NewScanner(file)
	var i int
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Trim(line, " \n\r\t")
		bits := strings.Split(line, " ")
		for j, k := range bits {
			k, err := strconv.Atoi(k)
			if err != nil {
			
			}
			schedule[j][i] = k
			
		}
		i++
	}
	
	return
}