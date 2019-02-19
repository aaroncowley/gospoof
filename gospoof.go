package main

import (
	_ "flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	_ "os/user"
	"sort"
	"strconv"
	_ "strings"

	"gopkg.in/src-d/go-git.v4"
)

func portRanges(nonSpoofPorts []string) (ranges string) {
	ranges = "1:"
	for i, port := range nonSpoofPorts {
		portNum, _ := strconv.Atoi(port)
		prevPort := 0
		nextPort := 0

		if i != 0 {
			prevPort, _ = strconv.Atoi(nonSpoofPorts[i-1])
		}

		if i != len(nonSpoofPorts)-1 {
			nextPort, _ = strconv.Atoi(nonSpoofPorts[i+1])
		}

		oneBelow := strconv.Itoa(portNum - 1)
		oneAbove := strconv.Itoa(portNum + 1)

		if portNum-1 == prevPort && portNum+1 != nextPort {
			ranges += oneAbove + ":"
		} else if portNum-1 == prevPort {
			continue
		} else if portNum+1 == nextPort {
			ranges += oneBelow + " "
		} else {
			ranges += (oneBelow + " " + oneAbove + ":")
		}
	}
	ranges += "65535"

	return ranges
}

func checkPorts(ports []string) {
	for _, port := range ports {
		//checks if ports are ints
		if _, err := strconv.ParseInt(port, 10, 16); err != nil {
			fmt.Println("Ports given to flag -nospoof invalid", err)
			os.Exit(1)
		}

		//checks for out of range
		portNum, _ := strconv.Atoi(port)
		if portNum >= 65535 || portNum <= 1 {
			fmt.Println("port number out of range (max is 65535, min is 1)")
			os.Exit(1)
		}
	}
}

func flushNatTable() {
	fmt.Println("flushing the iptable NAT table")
	cmd := exec.Command("iptables", "--table nat", "-F")
	err := cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
		os.Exit(1)
	}
}

func gitget() {
	fmt.Println("Cloning portspoof from github to /root/portspoof")

	//	Info("git clone https://github.com/drk1wi/portspoof.git /root/")

	_, err := git.PlainClone("/root/", false, &git.CloneOptions{
		URL:      "https://github.com/drk1wi/portspoof.git",
		Progress: os.Stdout,
	})
	if err != nil {
		log.Fatal(err)
	}

}

func main() {
	portArr := []string{"22", "23", "24", "25", "26", "334", "559", "8899", "21"}

	sort.Slice(portArr, func(i, j int) bool {
		numA, _ := strconv.Atoi(portArr[i])
		numB, _ := strconv.Atoi(portArr[j])
		return numA < numB
	})

	fmt.Printf("%v\n", portArr)

	rangey := portRanges(portArr)
	fmt.Println(rangey)

	if user.Uid != 0 {
		fmt.Println("gospoof must be ran as ROOT!")
		panic()
	}

	if _, err := os.Stat("/root/portspoof"); os.IsNotExist(err) {
		gitget()
	} else {
		Println("/root/portspoof exists, continuing execution")
	}

	//Command line flags
	legitPortHelp := `
	Pass ports that you would like untouched into this flag. Gospoof will create ranges
	around these ports so that they remain legitmate and will not be redirected.

	Port 22 is left by default to hopefully prevent an accidental machine lockout in
	the case of a miskey
	`
	legitPorts := flag.String("nospoof", "22", legitPortHelp)

	redirPortHelp := `
	change the port from the default (4444) to a new port

	Do not overlap address for service ports and portspoof, weirdness
	will occur
	`
	redirPort := flag.String("port", "4444", redirPortHelp)

	flag.Parse()

	//error checking in flags
	var noSpoofArr []string = strings.Split(*legitPorts, ' ')
	checkPorts(noSpoofArr)

	if strings.Contains(*legitPorts, *redirPort) {
		fmt.Println("Port overlap detected between Redirect port and Service port")
		os.Exit(1)
	}
}
