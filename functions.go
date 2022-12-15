package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"syscall"

	"golang.org/x/term"
)

func ProgUsage() bool {
	fmt.Println("Usage : nvgm [--server <PC Name>] [--user <Username>] [--help] [--usage] [--secure-mode]")
	os.Exit(1)
	return true
}

// This function get PE connection info from parameters or from STDIN
func GetPrismInfo() {

	// //Debug
	// MyPrism.PC = "pc.ntnx.fr"
	// MyPrism.Mode = "password"
	// MyPrism.User = "admin"
	// MyPrism.Password = "Nutanix/4u"
	// return //DEBUG

	// Define all parameters
	PC := flag.String("server", "", "Prism Element IP of FQDN")
	User := flag.String("user", "", "Prism Element User")
	help := flag.Bool("help", false, "Request usage")
	usage := flag.Bool("usage", false, "Request usage")
	secure := flag.Bool("secure-mode", false, "Request usage")
	flag.Parse()

	// Affect or request server value
	if *PC == string("") {
		fmt.Printf("Please enter Prism Central IP or FQDN : ")
		fmt.Scanln(&MyPrism.PC)
	} else {
		MyPrism.PC = *PC
	}

	if *help || *usage {
		ProgUsage()
	}

	// Affect or request user value
	if *User == string("") {
		fmt.Printf("Please enter Prism User : ")
		fmt.Scanln(&MyPrism.User)
	} else {
		MyPrism.User = *User
	}

	// Request password
	fmt.Printf("Please enter Prism password for " + MyPrism.User + " : ")
	tmp, _ := term.ReadPassword(int(syscall.Stdin))
	fmt.Println("")

	MyPrism.Password = string(tmp)

	// Define API call mode
	MyPrism.Mode = "password"

	// Deactivate SSL Check
	if *secure {
		ActivateSSLCheck(true)
	} else {
		ActivateSSLCheck(false)
	}

}

// =========== ActivateSSLCheck ===========
func ActivateSSLCheck(value bool) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: !value}

}

// =========== CheckErr ===========
// This function is will handle errors
func CheckErr(context string, err error) {
	if err != nil {
		fmt.Println("ERROR", context, " : ", err.Error())
		os.Exit(2)
	}
}
