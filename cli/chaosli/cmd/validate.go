// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2021 Datadog, Inc.

package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	goyaml "github.com/ghodss/yaml"

	"github.com/DataDog/chaos-controller/api/v1beta1"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "validate disruption config",
	Long:  `validates the yaml of the disruption for structure.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("path")
		return validatePath(path)
	},
	Run: ValidateDisruption,
}

func init() {
	validateCmd.Flags().String("path", "", "The path to the disruption file to be validated.")
}

func ValidateDisruption(cmd *cobra.Command, args []string) {
	path, _ := cmd.Flags().GetString("path")
	disruption, _ := DisruptionFromFile(path)

	//validate scope
	err := disruption.Spec.ValidateDisruptionSpec()
	if err != nil {
		log.Fatal(err)
	}
}

func validatePath(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("no path given, exiting")
	}
	return nil
}

func DisruptionFromFile(path string) (v1beta1.Disruption, error) {
	yaml, err := os.Open(path)
	if err != nil {
		return v1beta1.Disruption{}, err
	}

	yamlBytes, err := ioutil.ReadAll(yaml)
	if err != nil {
		return v1beta1.Disruption{}, err
	}

	parsedSpec := v1beta1.Disruption{}
	err = goyaml.Unmarshal(yamlBytes, &parsedSpec)

	if err != nil {
		return v1beta1.Disruption{}, err
	}
	return parsedSpec, nil
}
