package model

type AppConfig struct {
	TerraformStateBucketKey    string
	TerraformStateBucketSecret string
	TerraformStateBucketRegion string
	BuildNumber                string
}
