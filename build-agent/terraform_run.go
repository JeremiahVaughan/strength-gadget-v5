package main

import (
	"deploy/config"
	"deploy/service"
	"fmt"
	"os/exec"
)

func terraform(layerPath string, action string) error {
	initCmd := exec.Command(
		"terragrunt",
		"init",
		"-upgrade", // todo find a way to safely upgrade modules without doing it blindly everytime
		"-reconfigure",
		// todo find a way to commit checkin the .terraform.lock.hcl file somewhere so refreshing the Jenkins instance doesn't wipe out the lock file.
		"-backend-config=access_key="+config.AppConfig.TerraformStateBucketKey,
		"-backend-config=secret_key="+config.AppConfig.TerraformStateBucketSecret,
		"-backend-config=region="+config.AppConfig.TerraformStateBucketRegion,
	)
	initCmd.Dir = layerPath
	err := service.StreamCmd(initCmd)
	if err != nil {
		return fmt.Errorf("error when executing terragrunt init. Error: %v", err)
	}

	applyCmd := exec.Command("terragrunt", action, "-auto-approve")
	applyCmd.Dir = layerPath
	err = service.StreamCmd(applyCmd)
	if err != nil {
		return fmt.Errorf("error when executing terragrunt apply. Error: %v", err)
	}
	return nil
}
