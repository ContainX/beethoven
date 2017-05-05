package config

import (
	"path/filepath"
	"testing"
)

func TestDeprecatedFields(t *testing.T) {
	config, err := loadFromFile(filepath.Join("fixtures", "deprecated_fields.json"))
	if err != nil {
		t.Fatal(err)
	}

	if config.Marathon == nil {
		t.FailNow()
	}
	if config.SchedulerType != MarathonScheduler {
		t.Fail()
	}
	if config.Marathon.Username != "username" || config.Marathon.Password != "password" {
		t.Fail()
	}
}

func TestSwarmConfig(t *testing.T) {
	config, err := loadFromFile(filepath.Join("fixtures", "swarm_config.json"))
	if err != nil {
		t.Fatal(err)
	}

	if config.SchedulerType != SwarmScheduler {
		t.Error("Scheduler type was not Swarm")
	}

	if config.Swarm == nil {
		t.Fatal("Swarm configuration was nil")
	}

	if config.Swarm.Endpoint != "http://localhost:2222" {
		t.Error("Swarm endpoint did not match")
	}
}

func TestMarathonConfig(t *testing.T) {
	config, err := loadFromFile(filepath.Join("fixtures", "marathon_config.json"))
	if err != nil {
		t.Fatal(err)
	}

	if config.SchedulerType != MarathonScheduler {
		t.Error("Scheduler type was not Marathon")
	}

	if len(config.Marathon.Endpoints) != 1 {
		t.Error("Expected a single marathon endpoint")
	}

	if config.Marathon.Username != "username" {
		t.Error("Expected 'username' as value for username")
	}

	if config.Marathon.Password != "password" {
		t.Error("Expected 'password' as value for password")
	}

	if config.Marathon.ServiceId != "serviceId" {
		t.Error("Expected 'serviceId' as value for service_id")
	}
}

// Populates both Marathon and Swarm configuration and relies on scheduler_type
// to make decision during this test
func TestSchedulerTypeIsValid(t *testing.T) {
	config, err := loadFromFile(filepath.Join("fixtures", "scheduler_marathon.json"))
	if err != nil {
		t.Fatal(err)
	}

	if config.SchedulerType != MarathonScheduler {
		t.Error("Scheduler type was not Marathon")
	}

	config, err = loadFromFile(filepath.Join("fixtures", "scheduler_swarm.json"))
	if err != nil {
		t.Fatal(err)
	}

	if config.SchedulerType != SwarmScheduler {
		t.Error("Scheduler type was not Swarm")
	}
}
