package main

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"log"
	"os"
	"os/exec"
)

func main() {
	c := cron.New(cron.WithSeconds())

	err := ensureConfigFileExists()
	if err != nil {
		log.Fatalf("error, you must have a config file present at ./certbot/conf/cloudflare.ini else new certificates cannot be retrieved: %v", err)
	}

	// Attempting renewal during init to ensure any issues are caught right away rather, then waiting for the cron event
	err = attemptCertRenewal()
	if err != nil {
		log.Fatalf("error, when attempting cert renewal during init: %v", err)
	}

	_, err = c.AddFunc("@midnight", func() {
		err = attemptCertRenewal()
		if err != nil {
			log.Fatalf("error, when attempting cert renewal during cron job: %v", err)
		}
	})
	if err != nil {
		log.Fatalf("error, when scheduling watch: %v", err)
	}

	c.Start()

	// Cron Jobs forever
	select {}
}

func ensureConfigFileExists() error {
	_, fileCheck := os.Stat("/etc/letsencrypt/cloudflare.ini")
	if os.IsNotExist(fileCheck) {
		return fmt.Errorf("error, config file doesn't exist")
	}
	return nil
}

func attemptCertRenewal() error {
	present, err := isCertPresent()
	if err != nil {
		return fmt.Errorf("error, when checking if the cert requires renewal. Error: %v", err)
	}
	if present {
		err = renewCert()
		if err != nil {
			return fmt.Errorf("error, when attempting to renew cert. Error: %v", err)
		}
		err = gracefullyRefreshHaproxy()
		if err != nil {
			return fmt.Errorf("error, when attempting to gracefullyRefreshHaproxy() after renewCert(): %v", err)
		}
	} else {
		err = createNewCert()
		if err != nil {
			return fmt.Errorf("error, when generating a new cert: %v", err)
		}
		err = gracefullyRefreshHaproxy()
		if err != nil {
			return fmt.Errorf("error, when attempting to gracefullyRefreshHaproxy() after createNewCert(): %v", err)
		}
	}
	return nil
}

func createNewCert() error {
	cmd := exec.Command(
		"docker",
		"compose",
		"up",
		"-d",
		"certbot-new",
	)
	output, err := cmd.CombinedOutput()
	log.Println(string(output))
	if err != nil {
		return fmt.Errorf("error, when attempting to createNewCert(): %v", err)
	}
	return nil
}

func gracefullyRefreshHaproxy() error {
	cmd := exec.Command(
		"docker",
		"kill",
		"-s",
		"HUP",
		"home-proxy-v2-haproxy-1",
	)
	output, err := cmd.CombinedOutput()
	log.Println(string(output))
	if err != nil {
		return fmt.Errorf("error, when attempting to gracefullyRefreshHaproxy(): %v", err)
	}
	return nil
}

func isCertPresent() (bool, error) {
	dirPath := "/etc/letsencrypt/live/frii.day"
	_, dirCheck := os.Stat(dirPath)
	if os.IsNotExist(dirCheck) {
		return false, nil
	}
	certFilePath := fmt.Sprintf("%s/fullchain.pem", dirPath)
	_, fileCheck := os.Stat(certFilePath)
	if os.IsNotExist(fileCheck) {
		return false, nil
	}
	return true, nil
}

func renewCert() error {
	cmd := exec.Command(
		"docker",
		"compose",
		"up",
		"-d",
		"certbot-renew",
	)
	output, err := cmd.CombinedOutput()
	log.Println(string(output))
	if err != nil {
		return fmt.Errorf("error, when attempting to renewCert(): %v", err)
	}
	return nil
}
