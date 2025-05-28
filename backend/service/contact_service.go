package service

import (
	"contactmanager/models"
	"contactmanager/repository"
)

type ContactService struct {
	repo *repository.ContactRepository
}

func NewContactService(repo *repository.ContactRepository) *ContactService {
	return &ContactService{repo: repo}
}

func (s *ContactService) CreateContact(input *models.ContactInput) (*models.Contact, error) {
	return s.repo.Create(input)
}

func (s *ContactService) GetContact(id string) (*models.Contact, error) {
	return s.repo.GetByID(id)
}

func (s *ContactService) ListContacts(opts models.ListOptions) ([]*models.Contact, int, error) {
	if opts.Page <= 0 {
		opts.Page = 1
	}
	if opts.Limit <= 0 {
		opts.Limit = 20
	}
	if opts.Limit > 100 {
		opts.Limit = 100
	}

	return s.repo.List(opts)
}

func (s *ContactService) UpdateContact(id string, input *models.ContactInput) (*models.Contact, error) {
	return s.repo.Update(id, input)
}

func (s *ContactService) DeleteContact(id string) error {
	return s.repo.Delete(id)
}

func (s *ContactService) BulkCreateContacts(inputs []*models.ContactInput) ([]*models.Contact, []error) {
	var contacts []*models.Contact
	var errors []error

	for _, input := range inputs {
		contact, err := s.repo.Create(input)
		if err != nil {
			errors = append(errors, err)
			contacts = append(contacts, nil)
		} else {
			contacts = append(contacts, contact)
			errors = append(errors, nil)
		}
	}

	return contacts, errors
}
