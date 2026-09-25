package scategory

import (
	"context"
	"time"

	"github.com/gofrs/uuid"

	"github.com/fakel876/catalog-service/internal/app/entity"
	"github.com/fakel876/catalog-service/internal/app/repository"
	"github.com/fakel876/catalog-service/internal/app/service"
)

type srv struct {
	repoCategory repository.Category
	repoProduct  repository.Product
}

func NewService(repoCategory repository.Category, repoProduct repository.Product) service.Category {
	return &srv{
		repoCategory: repoCategory,
		repoProduct:  repoProduct,
	}
}

func (s *srv) Create(ctx context.Context, req entity.RequestCategoryCreate) (entity.Category, error) {
	existing, err := s.repoCategory.List(ctx, &req.Name)
	if err != nil {
		return entity.Category{}, err
	}
	if len(existing) > 0 {
		return entity.Category{}, entity.ErrAlreadyExists
	}

	now := time.Now()
	category := entity.Category{
		GUID:      uuid.Must(uuid.NewV4()),
		Name:      req.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repoCategory.Create(ctx, category); err != nil {
		return entity.Category{}, err
	}

	return category, nil
}

func (s *srv) GetByGUIDs(ctx context.Context, guids []uuid.UUID) ([]entity.Category, error) {
	return s.repoCategory.GetByGUIDs(ctx, guids)
}

func (s *srv) Update(ctx context.Context, guid uuid.UUID, req entity.RequestCategoryUpdate) (entity.Category, error) {
	categories, err := s.GetByGUIDs(ctx, []uuid.UUID{guid})
	if err != nil {
		return entity.Category{}, err
	}

	if len(categories) == 0 {
		return entity.Category{}, entity.ErrNotFound
	}

	existing, err := s.repoCategory.List(ctx, &req.Name)
	if err != nil {
		return entity.Category{}, err
	}

	for _, existingName := range existing {
		if existingName.GUID != guid {
			return entity.Category{}, entity.ErrAlreadyExists
		}
	}

	currentCategory := categories[0]
	currentCategory.Name = req.Name
	currentCategory.UpdatedAt = time.Now()

	if err := s.repoCategory.Update(ctx, currentCategory); err != nil {
		return entity.Category{}, err
	}

	return currentCategory, nil
}

func (s *srv) Delete(ctx context.Context, guid uuid.UUID) error {
	categories, err := s.GetByGUIDs(ctx, []uuid.UUID{guid})
	if err != nil {
		return err
	}

	if len(categories) == 0 {
		return entity.ErrNotFound
	}

	existing, err := s.repoProduct.List(ctx, nil, &guid)
	if err != nil {
		return err
	}

	if len(existing) > 0 {
		return entity.ErrCategoryHasProducts
	}

	return s.repoCategory.Delete(ctx, guid)
}

func (s *srv) List(ctx context.Context) ([]entity.Category, error) {
	return s.repoCategory.List(ctx, nil)
}
