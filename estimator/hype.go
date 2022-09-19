package estimator

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sajari/word2vec"
)

var (
	model     *word2vec.Model
	hypewords *[]string
	err       error
)

// Init is initialize db from main function
func Init() {
	model, err = LoadModel()
	if err != nil {
		log.Fatalf("error loading model: %v", err)
	}
	log.Printf("Hype estimator: [word size: %v, word dimension: %v]", model.Size(), model.Dim())
	hypewords = &[]string{
		"wtf",
		"WTF",
		"haha",
		"KEKW",
		"OMG",
		"LMFAO",
		"LMAO",
		"OMEGALUL",
		"Pog",
		"PogU",
		"LOL",
		"LUL",
		"PogChamp",
		"www",
		"ｗｗｗ",
		"ｗｗｗｗｗ",
		"うおおお",
		"おおおお",
		"ええええ",
		"ワロタ",
	}
}

// GetDB is called in models
func GetEstimator() *word2vec.Model {
	return model
}

func GetHypewords() *[]string {
	return hypewords
}

func LoadModel() (*word2vec.Model, error) {
	var model *word2vec.Model
	localfile := os.Getenv("LOCAL_MODEL")
	if len(localfile) != 0 {
		file, err := os.Open(localfile)
		if err != nil {
			return nil, err
		}
		model, err = word2vec.FromReader(file)
		if err != nil {
			return nil, err
		}
	} else {
		bucket := os.Getenv("S3_BUCKET_NAME")
		model_key := os.Getenv("S3_MODEL_KEY")
		if len(bucket) == 0 || len(model_key) == 0 {
			return nil, fmt.Errorf("AWS envs don't exist")
		}

		// Load the Shared AWS Configuration
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			return nil, err
		}

		// Create an Amazon S3 service client
		client := s3.NewFromConfig(cfg)

		output, err := client.GetObject(context.Background(), &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(model_key),
		})
		if err != nil {
			return nil, err
		}
		defer output.Body.Close()

		model, err = word2vec.FromReader(output.Body)
		if err != nil {
			return nil, err
		}
	}

	return model, nil
}
