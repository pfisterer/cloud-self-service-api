package storage

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Storage struct holds the GORM database connection.
type Storage struct {
	db *gorm.DB
}

// PolicyRule represents a DNS policy rule. It is the GORM model.
type PolicyRule struct {
	// GORM field tags are usually preferred for primary keys
	ID               int64     `gorm:"primaryKey" json:"id"`
	ZonePattern      string    `gorm:"type:varchar(255);uniqueIndex" json:"zone_pattern"`
	ZoneSoa          string    `gorm:"type:varchar(255);not null" json:"zone_soa"`
	TargetUserFilter string    `gorm:"type:varchar(255);not null" json:"target_user_filter"`
	Description      string    `gorm:"type:text;default:null" json:"description,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// NewStorage initializes the database connection and runs auto-migrations.
func NewStorage(dbType string, connectionString string) (*Storage, error) {
	var dialector gorm.Dialector
	var err error

	switch dbType {
	case "sqlite":
		dialector = sqlite.Open(connectionString)
	case "postgres":
		dialector = postgres.Open(connectionString)
	case "mysql":
		// Example: "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
		dialector = mysql.Open(connectionString)
	default:
		return nil, fmt.Errorf("storage.NewStorage: Unsupported database type: %s", dbType)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		// You may want to configure Logger/Tracing here for production
	})
	if err != nil {
		return nil, fmt.Errorf("storage.NewStorage: Failed to connect to %s database: %w", dbType, err)
	}

	// AutoMigrate creates tables/columns based on the model if they don't exist
	err = db.AutoMigrate(&PolicyRule{})
	if err != nil {
		return nil, fmt.Errorf("storage.NewStorage: Failed to auto-migrate database: %w", err)
	}

	return &Storage{db: db}, nil
}

// -- Insert dummy data function (optional) --
func (s *Storage) PolicyInsertDummyData() error {
	dummyRules := []PolicyRule{
		{ZonePattern: "%u.users.dhbw.cloud", ZoneSoa: "users.dhbw.cloud", TargetUserFilter: "*@dhbw.de", Description: "Automatic personal zones for DHBW users", CreatedAt: time.Now().Add(-24 * time.Hour)},
		{ZonePattern: "project.dhbw.cloud", ZoneSoa: "project.dhbw.cloud", TargetUserFilter: "*@dhbw.de", Description: "All DHBW users can manage a common project zone", CreatedAt: time.Now().Add(-24 * time.Hour)},
		{ZonePattern: "%u.cloud.uni-luebeck.de", ZoneSoa: "cloud.uni-luebeck.de", TargetUserFilter: "*@uni-luebeck.de", Description: "All Uni-Luebeck users can create subdomains", CreatedAt: time.Now().Add(-24 * time.Hour)},
	}

	for _, rule := range dummyRules {
		_, err := s.PolicyCreate(&rule)
		if err != nil {
			return fmt.Errorf("storage.InsertDummyData: Failed to insert dummy data: %w", err)
		}
	}
	return nil
}

// --- CRUD Operations for PolicyRule ---

// PolicyCreate inserts a new PolicyRule into the database.
func (s *Storage) PolicyCreate(rule *PolicyRule) (*PolicyRule, error) {
	// Set creation timestamp manually if not using GORM's default fields
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = time.Now()
	}

	result := s.db.Create(rule)
	if result.Error != nil {
		// Handle potential unique constraint violation (e.g., if ZonePattern is marked unique)
		return nil, fmt.Errorf("storage.Create: Failed to create rule: %w", result.Error)
	}
	return rule, nil
}

// PolicyGetAll retrieves all PolicyRules from the database.
func (s *Storage) PolicyGetAll() ([]PolicyRule, error) {
	var rules []PolicyRule
	// Order by ID or Creation Time for consistent results
	result := s.db.Order("id asc").Find(&rules)
	if result.Error != nil {
		return nil, fmt.Errorf("storage.GetAll: Failed to retrieve rules: %w", result.Error)
	}
	return rules, nil
}

// PolicyGetByID retrieves a single PolicyRule by its ID.
func (s *Storage) PolicyGetByID(id int64) (*PolicyRule, error) {
	var rule PolicyRule
	result := s.db.First(&rule, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound // Return GORM's specific error for Not Found
		}
		return nil, fmt.Errorf("storage.GetByID: Failed to retrieve rule %d: %w", id, result.Error)
	}
	return &rule, nil
}

// PolicyUpdate modifies an existing PolicyRule.
// The rule parameter should contain the ID of the rule to update and the new values.
func (s *Storage) PolicyUpdate(rule *PolicyRule) (*PolicyRule, error) {
	// GORM will use the primary key (ID) of the struct to determine which record to update.
	// We use Select to specify only the fields we allow the user to modify.
	result := s.db.Model(rule).Select("ZonePattern", "TargetUserFilter", "Description").Updates(rule)

	if result.Error != nil {
		return nil, fmt.Errorf("storage.Update: Failed to update rule %d: %w", rule.ID, result.Error)
	}

	if result.RowsAffected == 0 {
		// Double check if the record was actually found and updated
		// Fetch the record again to return a complete, updated object (optional but safer)
		if _, err := s.PolicyGetByID(rule.ID); err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, gorm.ErrRecordNotFound
			}
			return nil, fmt.Errorf("storage.Update: Rule %d was not found after attempted update: %w", rule.ID, err)
		}
	}

	// Return the rule object, which now has the updated fields and original ID/timestamps.
	return rule, nil
}

// PolicyDelete removes a PolicyRule from the database by its ID.
func (s *Storage) PolicyDelete(id int64) error {
	// Delete the record matching the ID
	result := s.db.Delete(&PolicyRule{}, id)

	if result.Error != nil {
		return fmt.Errorf("storage.Delete: Failed to delete rule %d: %w", id, result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // Indicate that no record with that ID was found
	}

	return nil
}
