package iam

import (
	"fmt"
	"strings"
	"time"

	"github.com/els/backend/internal/domain/shared"
	"github.com/els/backend/internal/domain/shared/vo"
	"github.com/els/backend/internal/utils/timex"
)

type AccountID struct{ vo.ID }

func NewAccountID() AccountID { return AccountID{ID: vo.NewID()} }

const DefaultNativeLanguage = "Russian"

type Account struct {
	id               AccountID
	email            vo.Email
	name             vo.PersonName
	pictureURL       string
	englishLevel     string
	aboutMe          string
	nativeLanguage   string
	showTranslations bool
	status           AccountStatus
	timestamps       vo.Timestamps
}

type NewAccountParams struct {
	ID               AccountID
	Email            string
	FirstName        string
	LastName         string
	PictureURL       string
	EnglishLevel     string
	AboutMe          string
	NativeLanguage   string
	ShowTranslations bool
	Status           AccountStatus
	Timestamps       vo.Timestamps
}

func NewAccount(p NewAccountParams) (*Account, error) {
	email, emailErr := vo.NewEmail(p.Email)
	name, nameErr := vo.NewPersonName(p.FirstName, p.LastName)

	var errs []error
	if p.ID.IsZero() {
		errs = append(errs, fmt.Errorf("account.id: must not be zero"))
	}
	if emailErr != nil {
		errs = append(errs, fmt.Errorf("account.email: %w", emailErr))
	}
	if nameErr != nil {
		errs = append(errs, fmt.Errorf("account.name: %w", nameErr))
	}
	if !p.Status.IsValid() {
		errs = append(errs, fmt.Errorf("account.status: invalid %q", p.Status))
	}
	if _, err := vo.NewTimestamps(p.Timestamps.CreatedAt(), p.Timestamps.UpdatedAt()); err != nil {
		errs = append(errs, fmt.Errorf("account.timestamps: %w", err))
	}
	if err := shared.Validation(errs...); err != nil {
		return nil, err
	}

	nativeLanguage := strings.TrimSpace(p.NativeLanguage)
	if nativeLanguage == "" {
		nativeLanguage = DefaultNativeLanguage
	}

	return &Account{
		id:               p.ID,
		email:            email,
		name:             name,
		pictureURL:       strings.TrimSpace(p.PictureURL),
		englishLevel:     strings.TrimSpace(p.EnglishLevel),
		aboutMe:          strings.TrimSpace(p.AboutMe),
		nativeLanguage:   nativeLanguage,
		showTranslations: p.ShowTranslations,
		status:           p.Status,
		timestamps:       p.Timestamps,
	}, nil
}

func NewPendingAccountNow(id AccountID, email, firstName, lastName string) (*Account, error) {
	timestamps, err := vo.NewCurrentTimestamps()
	if err != nil {
		return nil, shared.Validation(fmt.Errorf("account.timestamps: %w", err))
	}
	return NewAccount(NewAccountParams{
		ID:               id,
		Email:            email,
		FirstName:        firstName,
		LastName:         lastName,
		ShowTranslations: true,
		Status:           AccountStatusPendingPassword,
		Timestamps:       timestamps,
	})
}

func (a *Account) ID() AccountID          { return a.id }
func (a *Account) Email() vo.Email        { return a.email }
func (a *Account) Name() vo.PersonName    { return a.name }
func (a *Account) PictureURL() string     { return a.pictureURL }
func (a *Account) EnglishLevel() string   { return a.englishLevel }
func (a *Account) AboutMe() string        { return a.aboutMe }
func (a *Account) NativeLanguage() string { return a.nativeLanguage }
func (a *Account) ShowTranslations() bool { return a.showTranslations }
func (a *Account) Status() AccountStatus  { return a.status }
func (a *Account) IsActive() bool         { return a.status == AccountStatusActive }
func (a *Account) CreatedAt() time.Time   { return a.timestamps.CreatedAt() }
func (a *Account) UpdatedAt() time.Time   { return a.timestamps.UpdatedAt() }

func (a *Account) EnsureCanLogin() error {
	switch a.status {
	case AccountStatusActive:
		return nil
	case AccountStatusPendingPassword:
		return ErrAccountPending
	case AccountStatusBlocked:
		return ErrAccountBlocked
	case AccountStatusNoAuth:
		return ErrAccountBlocked
	}
	return ErrAccountBlocked
}

func (a *Account) Activate() error {
	if a.status == AccountStatusActive {
		return nil
	}
	if a.status != AccountStatusPendingPassword {
		return fmt.Errorf("%w: cannot activate account in status %s", shared.ErrConflict, a.status)
	}
	return a.transitionTo(AccountStatusActive)
}

func (a *Account) Block() error {
	if a.status == AccountStatusBlocked {
		return nil
	}
	return a.transitionTo(AccountStatusBlocked)
}

func (a *Account) Unblock() error {
	if a.status != AccountStatusBlocked {
		return nil
	}
	return a.transitionTo(AccountStatusActive)
}

func (a *Account) Rename(name vo.PersonName) error {
	if name.IsZero() {
		return shared.Validation(fmt.Errorf("account.name: must not be empty"))
	}
	timestamps, err := a.timestamps.Touch(timex.Now())
	if err != nil {
		return shared.Validation(fmt.Errorf("account.timestamps: %w", err))
	}
	a.name = name
	a.timestamps = timestamps
	return nil
}

func (a *Account) UpdateProfile(name vo.PersonName, englishLevel, aboutMe, nativeLanguage string, showTranslations bool) error {
	if name.IsZero() {
		return shared.Validation(fmt.Errorf("account.name: must not be empty"))
	}
	englishLevel = strings.TrimSpace(englishLevel)
	aboutMe = strings.TrimSpace(aboutMe)
	nativeLanguage = strings.TrimSpace(nativeLanguage)
	if len(englishLevel) > 100 {
		return shared.Validation(fmt.Errorf("account.english_level: too long"))
	}
	if len(aboutMe) > 2000 {
		return shared.Validation(fmt.Errorf("account.about_me: too long"))
	}
	if len(nativeLanguage) > 100 {
		return shared.Validation(fmt.Errorf("account.native_language: too long"))
	}
	if nativeLanguage == "" {
		nativeLanguage = DefaultNativeLanguage
	}
	timestamps, err := a.timestamps.Touch(timex.Now())
	if err != nil {
		return shared.Validation(fmt.Errorf("account.timestamps: %w", err))
	}
	a.name = name
	a.englishLevel = englishLevel
	a.aboutMe = aboutMe
	a.nativeLanguage = nativeLanguage
	a.showTranslations = showTranslations
	a.timestamps = timestamps
	return nil
}

func (a *Account) ChangePictureURL(url string) error {
	timestamps, err := a.timestamps.Touch(timex.Now())
	if err != nil {
		return shared.Validation(fmt.Errorf("account.timestamps: %w", err))
	}
	a.pictureURL = strings.TrimSpace(url)
	a.timestamps = timestamps
	return nil
}

func (a *Account) ChangeEmail(email vo.Email) error {
	if email.IsZero() {
		return shared.Validation(fmt.Errorf("account.email: must not be empty"))
	}
	if a.email.String() == email.String() {
		return nil
	}
	timestamps, err := a.timestamps.Touch(timex.Now())
	if err != nil {
		return shared.Validation(fmt.Errorf("account.timestamps: %w", err))
	}
	a.email = email
	a.timestamps = timestamps
	return nil
}

func (a *Account) transitionTo(s AccountStatus) error {
	timestamps, err := a.timestamps.Touch(timex.Now())
	if err != nil {
		return shared.Validation(fmt.Errorf("account.timestamps: %w", err))
	}
	a.status = s
	a.timestamps = timestamps
	return nil
}
