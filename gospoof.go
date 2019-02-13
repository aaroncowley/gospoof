package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"request"
	"strconv"
	"strings"

	"gopkg.in/src-d/go-git.v4"
)

func checkPorts(ports []string) {
	for _, port := range ports {
		if _, err := strconv.ParseInt(port, 10, 64); err != nil {
			fmt.Println("Ports given to flag -nospoof invalid")
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

	Info("git clone", "https://github.com/drk1wi/portspoof.git", "/root/")

	_, err := git.PlainClone("/root/", false, &git.CloneOptions{
		URL:      "https://github.com/drk1wi/portspoof.git",
		Progress: os.Stdout,
	})

}

func main() {
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
