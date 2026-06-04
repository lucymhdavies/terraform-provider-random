// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestResourceFailure_ShouldFailByPercentage_FiftyPercent_Statistical(t *testing.T) {
	if os.Getenv("TF_ACC_RANDOM_FAILURE_STAT_TEST") == "" {
		t.Skip("set TF_ACC_RANDOM_FAILURE_STAT_TEST=1 to run statistical random_failure test")
	}

	const runCount = 100
	const lowerBoundPercent = 30
	const upperBoundPercent = 70
	const targetPercentage = 50

	failureCount := 0

	for range runCount {
		failed, err := shouldFailByPercentage(targetPercentage)
		if err != nil {
			t.Fatalf("unexpected error generating create failure outcome: %v", err)
		}

		if failed {
			failureCount++
		}
	}

	failurePercent := failureCount * 100 / runCount

	if failurePercent < lowerBoundPercent || failurePercent > upperBoundPercent {
		t.Fatalf(
			"observed failure rate %d%% (%d/%d) was outside expected range %d%%-%d%% for create_failure_percentage=%d",
			failurePercent,
			failureCount,
			runCount,
			lowerBoundPercent,
			upperBoundPercent,
			targetPercentage,
		)
	}

	t.Log(fmt.Sprintf(
		"observed failure rate %d%% (%d/%d) within expected range %d%%-%d%%",
		failurePercent,
		failureCount,
		runCount,
		lowerBoundPercent,
		upperBoundPercent,
	))
}

func TestAccResourceFailure_Defaults(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_failure" "test" {}`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("random_failure.test", tfjsonpath.New("create_failure_percentage"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue("random_failure.test", tfjsonpath.New("update_failure_percentage"), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue("random_failure.test", tfjsonpath.New("destroy_failure_percentage"), knownvalue.Int64Exact(0)),
				},
			},
		},
	})
}

func TestAccResourceFailure_Create_FailsAtHundredPercent(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_failure" "test" {
					create_failure_percentage = 100
				}`,
				ExpectError: regexp.MustCompile(`(?s).*create.*failed.*`),
			},
		},
	})
}

func TestAccResourceFailure_Update_FailsAtHundredPercent(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_failure" "test" {
					create_failure_percentage  = 0
					update_failure_percentage  = 0
					destroy_failure_percentage = 0
				}`,
			},
			{
				Config: `resource "random_failure" "test" {
					create_failure_percentage  = 0
					update_failure_percentage  = 100
					destroy_failure_percentage = 0
				}`,
				ExpectError: regexp.MustCompile(`(?s).*update.*failed.*`),
			},
		},
	})
}

func TestAccResourceFailure_Destroy_FailsAtHundredPercent(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_failure" "test" {
					create_failure_percentage  = 0
					update_failure_percentage  = 0
					destroy_failure_percentage = 100
				}`,
			},
			{
				Config: `resource "random_failure" "test" {
					create_failure_percentage  = 0
					update_failure_percentage  = 0
					destroy_failure_percentage = 100
				}`,
				Destroy:     true,
				ExpectError: regexp.MustCompile(`(?s).*destroy.*failed.*`),
			},
			{
				Config: `resource "random_failure" "test" {
					create_failure_percentage  = 0
					update_failure_percentage  = 0
					destroy_failure_percentage = 0
				}`,
			},
		},
	})
}

func TestAccResourceFailure_PercentageValidationErrors(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: protoV5ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "random_failure" "invalid_create_min" {
					create_failure_percentage = -1
				}`,
				ExpectError: regexp.MustCompile(`(?s).*Attribute create_failure_percentage value must be at least 0.*`),
			},
			{
				Config: `resource "random_failure" "invalid_update_max" {
					update_failure_percentage = 101
				}`,
				ExpectError: regexp.MustCompile(`(?s).*Attribute update_failure_percentage value must be at most 100.*`),
			},
			{
				Config: `resource "random_failure" "invalid_destroy_min" {
					destroy_failure_percentage = -1
				}`,
				ExpectError: regexp.MustCompile(`(?s).*Attribute destroy_failure_percentage value must be at least 0.*`),
			},
		},
	})
}
