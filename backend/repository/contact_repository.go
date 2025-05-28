package repository

import (
	"contactmanager/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type ContactRepository struct {
	db *sql.DB
}

func NewContactRepository(db *sql.DB) *ContactRepository {
	return &ContactRepository{db: db}
}

func (r *ContactRepository) CreateTable() error {
	query := `
    CREATE TABLE IF NOT EXISTS contacts (
        id TEXT PRIMARY KEY,
        first_name TEXT NOT NULL,
        last_name TEXT NOT NULL,
        email TEXT UNIQUE NOT NULL,
        phone_number TEXT,
        address_street TEXT,
        address_city TEXT,
        address_state TEXT,
        address_postal_code TEXT,
        address_country TEXT,
        company TEXT,
        job_title TEXT,
        tags TEXT,
        notes TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    
    CREATE INDEX IF NOT EXISTS idx_contacts_email ON contacts(email);
    CREATE INDEX IF NOT EXISTS idx_contacts_name ON contacts(first_name, last_name);
    CREATE INDEX IF NOT EXISTS idx_contacts_company ON contacts(company);
    `
	_, err := r.db.Exec(query)
	return err
}

func (r *ContactRepository) Create(input *models.ContactInput) (*models.Contact, error) {
	id := uuid.New().String()
	now := time.Now()

	// Serialize tags to JSON
	tagsJSON, _ := json.Marshal(input.Tags)

	contact := &models.Contact{
		ID:        id,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Handle optional fields
	if input.PhoneNumber != nil {
		contact.PhoneNumber = input.PhoneNumber
	}
	if input.Company != nil {
		contact.Company = input.Company
	}
	if input.JobTitle != nil {
		contact.JobTitle = input.JobTitle
	}
	if input.Notes != nil {
		contact.Notes = input.Notes
	}
	if input.Tags != nil {
		contact.Tags = input.Tags
	}
	if input.Address != nil {
		contact.Address = input.Address
	}

	query := `
    INSERT INTO contacts (
        id, first_name, last_name, email, phone_number, 
        address_street, address_city, address_state, address_postal_code, address_country,
        company, job_title, tags, notes, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

	var addressStreet, addressCity, addressState, addressPostalCode, addressCountry *string
	if input.Address != nil {
		addressStreet = input.Address.Street
		addressCity = input.Address.City
		addressState = input.Address.State
		addressPostalCode = input.Address.PostalCode
		addressCountry = input.Address.Country
	}

	_, err := r.db.Exec(query,
		id, input.FirstName, input.LastName, input.Email, input.PhoneNumber,
		addressStreet, addressCity, addressState, addressPostalCode, addressCountry,
		input.Company, input.JobTitle, string(tagsJSON), input.Notes, now, now,
	)

	if err != nil {
		return nil, err
	}

	return contact, nil
}

func (r *ContactRepository) GetByID(id string) (*models.Contact, error) {
	query := `
    SELECT id, first_name, last_name, email, phone_number,
           address_street, address_city, address_state, address_postal_code, address_country,
           company, job_title, tags, notes, created_at, updated_at
    FROM contacts WHERE id = ?
    `

	var contact models.Contact
	var tagsJSON string
	var addressStreet, addressCity, addressState, addressPostalCode, addressCountry sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&contact.ID, &contact.FirstName, &contact.LastName, &contact.Email, &contact.PhoneNumber,
		&addressStreet, &addressCity, &addressState, &addressPostalCode, &addressCountry,
		&contact.Company, &contact.JobTitle, &tagsJSON, &contact.Notes, &contact.CreatedAt, &contact.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Deserialize tags
	if tagsJSON != "" {
		json.Unmarshal([]byte(tagsJSON), &contact.Tags)
	}

	// Build address if any field is present
	if addressStreet.Valid || addressCity.Valid || addressState.Valid || addressPostalCode.Valid || addressCountry.Valid {
		contact.Address = &models.Address{}
		if addressStreet.Valid {
			contact.Address.Street = &addressStreet.String
		}
		if addressCity.Valid {
			contact.Address.City = &addressCity.String
		}
		if addressState.Valid {
			contact.Address.State = &addressState.String
		}
		if addressPostalCode.Valid {
			contact.Address.PostalCode = &addressPostalCode.String
		}
		if addressCountry.Valid {
			contact.Address.Country = &addressCountry.String
		}
	}

	return &contact, nil
}

func (r *ContactRepository) List(opts models.ListOptions) ([]*models.Contact, int, error) {
	// Build WHERE clause
	where := "1=1"
	args := []interface{}{}

	if opts.Search != "" {
		where += " AND (first_name LIKE ? OR last_name LIKE ? OR email LIKE ?)"
		searchTerm := "%" + opts.Search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
	}

	if opts.Tag != "" {
		where += " AND tags LIKE ?"
		args = append(args, "%\""+opts.Tag+"\"%")
	}

	// Count total records
	countQuery := "SELECT COUNT(*) FROM contacts WHERE " + where
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (opts.Page - 1) * opts.Limit
	query := fmt.Sprintf(`
        SELECT id, first_name, last_name, email, phone_number,
               address_street, address_city, address_state, address_postal_code, address_country,
               company, job_title, tags, notes, created_at, updated_at
        FROM contacts WHERE %s
        ORDER BY created_at DESC
        LIMIT ? OFFSET ?
    `, where)

	args = append(args, opts.Limit, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var contacts []*models.Contact
	for rows.Next() {
		var contact models.Contact
		var tagsJSON string
		var addressStreet, addressCity, addressState, addressPostalCode, addressCountry sql.NullString

		err := rows.Scan(
			&contact.ID, &contact.FirstName, &contact.LastName, &contact.Email, &contact.PhoneNumber,
			&addressStreet, &addressCity, &addressState, &addressPostalCode, &addressCountry,
			&contact.Company, &contact.JobTitle, &tagsJSON, &contact.Notes, &contact.CreatedAt, &contact.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		// Deserialize tags
		if tagsJSON != "" {
			json.Unmarshal([]byte(tagsJSON), &contact.Tags)
		}

		// Build address if any field is present
		if addressStreet.Valid || addressCity.Valid || addressState.Valid || addressPostalCode.Valid || addressCountry.Valid {
			contact.Address = &models.Address{}
			if addressStreet.Valid {
				contact.Address.Street = &addressStreet.String
			}
			if addressCity.Valid {
				contact.Address.City = &addressCity.String
			}
			if addressState.Valid {
				contact.Address.State = &addressState.String
			}
			if addressPostalCode.Valid {
				contact.Address.PostalCode = &addressPostalCode.String
			}
			if addressCountry.Valid {
				contact.Address.Country = &addressCountry.String
			}
		}

		contacts = append(contacts, &contact)
	}

	return contacts, total, nil
}

func (r *ContactRepository) Update(id string, input *models.ContactInput) (*models.Contact, error) {
	now := time.Now()
	tagsJSON, _ := json.Marshal(input.Tags)

	var addressStreet, addressCity, addressState, addressPostalCode, addressCountry *string
	if input.Address != nil {
		addressStreet = input.Address.Street
		addressCity = input.Address.City
		addressState = input.Address.State
		addressPostalCode = input.Address.PostalCode
		addressCountry = input.Address.Country
	}

	query := `
    UPDATE contacts SET 
        first_name = ?, last_name = ?, email = ?, phone_number = ?,
        address_street = ?, address_city = ?, address_state = ?, address_postal_code = ?, address_country = ?,
        company = ?, job_title = ?, tags = ?, notes = ?, updated_at = ?
    WHERE id = ?
    `

	_, err := r.db.Exec(query,
		input.FirstName, input.LastName, input.Email, input.PhoneNumber,
		addressStreet, addressCity, addressState, addressPostalCode, addressCountry,
		input.Company, input.JobTitle, string(tagsJSON), input.Notes, now, id,
	)

	if err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *ContactRepository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM contacts WHERE id = ?", id)
	return err
}
