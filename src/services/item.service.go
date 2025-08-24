package services

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ItemService struct {
	UploadRepository      repositories.UploadRepository
	ItemRepository        repositories.ItemRepository
	ItemHistoryRepository repositories.ItemHistoryRepository
}

func NewItemService(uploadRepo repositories.UploadRepository, itemRepo repositories.ItemRepository, itemHistoryRepo repositories.ItemHistoryRepository) *ItemService {
	return &ItemService{
		UploadRepository:      uploadRepo,
		ItemRepository:        itemRepo,
		ItemHistoryRepository: itemHistoryRepo,
	}
}

// ==============================
// Reads
// ==============================

func (s *ItemService) GetAllItems() ([]models.ResponseGetItem, error) {
	items, err := s.ItemRepository.FindAll(nil)
	if err != nil {
		return nil, err
	}

	resp := make([]models.ResponseGetItem, 0, len(items))
	for _, it := range items {
		resp = append(resp, models.ResponseGetItem{
			ID:            it.ID,
			Name:          it.Name,
			Code:          it.Code,
			Price:         it.Price,
			ImageID:       it.ImageID,
			Image:         it.Image,
			CategoryID:    it.CategoryID,
			Category:      it.Category,
			UoMID:         it.UoMID,
			UoM:           it.UoM,
			Description:   it.Description,
			Batch: 					it.Batch,
			ExpiredAt: 		it.ExpiredAt,
			Stock:         it.Stock,
			LowStock:      it.LowStock,
			ItemHistories: it.ItemHistories,
			CreatedAt:     it.CreatedAt,
			UpdatedAt:     it.UpdatedAt,
			DeletedAt:     it.DeletedAt,
		})
	}
	return resp, nil
}

func (s *ItemService) GetAllItemsPaginated(req *models.PaginationRequest, userInfo *models.User) (*models.ItemPaginatedResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Status == "" {
		req.Status = "active"
	}

	items, totalCount, err := s.ItemRepository.FindAllPaginated(nil, req)
	if err != nil {
		return nil, err
	}

	itemsResp := make([]models.ResponseGetItem, 0, len(items))
	for _, it := range items {
		itemsResp = append(itemsResp, models.ResponseGetItem{
			ID:            it.ID,
			Name:          it.Name,
			Code:          it.Code,
			Price:         it.Price,
			ImageID:       it.ImageID,
			Image:         it.Image,
			CategoryID:    it.CategoryID,
			Category:      it.Category,
			UoMID:         it.UoMID,
			UoM:           it.UoM,
			Description:   it.Description,
			Batch: 					it.Batch,
			ExpiredAt: 		it.ExpiredAt,
			Stock:         it.Stock,
			LowStock:      it.LowStock,
			ItemHistories: it.ItemHistories,
			CreatedAt:     it.CreatedAt,
			UpdatedAt:     it.UpdatedAt,
			DeletedAt:     it.DeletedAt,
		})
	}

	totalPages := int((totalCount + int64(req.Limit) - 1) / int64(req.Limit))
	p := models.PaginationResponse{
		CurrentPage:  req.Page,
		PerPage:      req.Limit,
		TotalPages:   totalPages,
		TotalRecords: totalCount,
		HasNext:      req.Page < totalPages,
		HasPrev:      req.Page > 1,
	}

	return &models.ItemPaginatedResponse{Data: itemsResp, Pagination: p}, nil
}

func (s *ItemService) GetItemByID(itemId string) (*models.ResponseGetItem, error) {
	it, err := s.ItemRepository.FindById(nil, itemId, false)
	if err != nil {
		return nil, err
	}
	return &models.ResponseGetItem{
			ID:            it.ID,
			Name:          it.Name,
			Code:          it.Code,
			Price:         it.Price,
			ImageID:       it.ImageID,
			Image:         it.Image,
			CategoryID:    it.CategoryID,
			Category:      it.Category,
			UoMID:         it.UoMID,
			UoM:           it.UoM,
			Description:   it.Description,
			Batch: 					it.Batch,
			ExpiredAt: 		it.ExpiredAt,
			Stock:         it.Stock,
			LowStock:      it.LowStock,
			ItemHistories: it.ItemHistories,
	}, nil
}

// ==============================
// Mutations (transaction-aware; pake configs.DB.Begin())
// ==============================

func (s *ItemService) CreateItem(req *models.ItemCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Item, error) {
	var newImageUUIDStr string

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			if newImageUUIDStr != "" {
				helpers.DeleteLocalFileImmediate(newImageUUIDStr)
			}
		}
	}()

	// cek unik nama
	if _, err := s.ItemRepository.FindByName(tx, req.Name); err == nil {
		tx.Rollback()
		return nil, errors.New("item already exists")
	} else if err != nil && err != repositories.ErrItemNotFound {
		tx.Rollback()
		return nil, fmt.Errorf("error checking item: %w", err)
	}

	newItem := &models.Item{
		ID:          uuid.New(),
		Name:        req.Name,
		Code:        req.Code,
		Price:       req.Price,
		Stock:       req.Stock,
		LowStock:    req.LowStock,
		CategoryID:  req.CategoryID,
		UoMID:       req.UoMID,
		Description: req.Description,
		Batch:       req.Batch,
		ExpiredAt:   req.ExpiredAt,
	}

	// handle image (optional)
	if file, err := ctx.FormFile("image"); err == nil && file != nil {
		var err error
		newImageUUIDStr, err = helpers.SaveFile(ctx, file, "items")
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to save image: %w", err)
		}
		imgUUID, err := uuid.Parse(newImageUUIDStr)
		if err != nil {
			tx.Rollback()
			helpers.DeleteLocalFileImmediate(newImageUUIDStr)
			return nil, fmt.Errorf("invalid image UUID: %w", err)
		}
		newItem.ImageID = &imgUUID
	}

	created, err := s.ItemRepository.Insert(tx, newItem)
	if err != nil {
		tx.Rollback()
		if newImageUUIDStr != "" {
			helpers.DeleteLocalFileImmediate(newImageUUIDStr)
		}
		return nil, err
	}

	// histories (atomic bareng create)
	if _, err := s.ItemHistoryRepository.Insert(tx, &models.ItemHistory{
		ID:           uuid.New(),
		ItemID:       created.ID,
		ChangeType:   "create_price",
		OldPrice:     0,
		NewPrice:     created.Price,
		CurrentPrice: created.Price,
		CreatedBy:    &userInfo.ID,
		UpdatedBy:    &userInfo.ID,
		Description:  "Initial price set to " + strconv.Itoa(created.Price),
	}); err != nil {
		tx.Rollback()
		if newImageUUIDStr != "" {
			helpers.DeleteLocalFileImmediate(newImageUUIDStr)
		}
		return nil, err
	}
	if _, err := s.ItemHistoryRepository.Insert(tx, &models.ItemHistory{
		ID:           uuid.New(),
		ItemID:       created.ID,
		ChangeType:   "create_stock",
		OldStock:     0,
		NewStock:     created.Stock,
		CurrentStock: created.Stock,
		CreatedBy:    &userInfo.ID,
		UpdatedBy:    &userInfo.ID,
		Description:  "Initial stock set to " + strconv.Itoa(created.Stock),
	}); err != nil {
		tx.Rollback()
		if newImageUUIDStr != "" {
			helpers.DeleteLocalFileImmediate(newImageUUIDStr)
		}
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		if newImageUUIDStr != "" {
			helpers.DeleteLocalFileImmediate(newImageUUIDStr)
		}
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	inserted, err := s.ItemRepository.FindById(nil, created.ID.String(), false)
	if err != nil {
		log.Printf("Warning: Item created but failed to fetch created data: %v", err)
		return created, nil
	}
	return inserted, nil
}

func (s *ItemService) UpdateItem(req *models.ItemUpdateRequest, itemID string, ctx *fiber.Ctx, userInfo *models.User) (*models.Item, error) {
	var oldImageIDToDelete *uuid.UUID
	var newImageUUIDStr string

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			if newImageUUIDStr != "" {
				helpers.DeleteLocalFileImmediate(newImageUUIDStr)
			}
		}
	}()

	item, err := s.ItemRepository.FindById(tx, itemID, false)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if item == nil {
		tx.Rollback()
		return nil, errors.New("item not found")
	}

	// patch fields
	if req.Name != "" {
		item.Name = req.Name
	}
	if req.Code != "" {
		item.Code = req.Code
	}
	if req.LowStock != 0 {
		item.LowStock = req.LowStock
	}
	if req.CategoryID != uuid.Nil {
		item.CategoryID = req.CategoryID
	}
	if req.UoMID != uuid.Nil {
		item.UoMID = req.UoMID
	}
	if req.Description != "" {
		item.Description = req.Description
	}
	if req.Batch != 0 {
		item.Batch = req.Batch
	}
	if !req.ExpiredAt.IsZero() {
		item.ExpiredAt = req.ExpiredAt
	}

	// optional image replacement
	if file, err := ctx.FormFile("image"); err == nil && file != nil {
		newImageUUIDStr, err = helpers.SaveFile(ctx, file, "items")
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to save image: %w", err)
		}
		newImageID, err := uuid.Parse(newImageUUIDStr)
		if err != nil {
			tx.Rollback()
			helpers.DeleteLocalFileImmediate(newImageUUIDStr)
			return nil, fmt.Errorf("invalid image UUID: %w", err)
		}
		// simpan image lama untuk dihapus setelah commit
		if item.ImageID != nil {
			tmp := *item.ImageID
			oldImageIDToDelete = &tmp
		}
		item.Image = nil
		item.ImageID = &newImageID
	}

	updated, err := s.ItemRepository.Update(tx, item)
	if err != nil {
		tx.Rollback()
		if newImageUUIDStr != "" {
			helpers.DeleteLocalFileImmediate(newImageUUIDStr)
		}
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		if newImageUUIDStr != "" {
			helpers.DeleteLocalFileImmediate(newImageUUIDStr)
		}
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// setelah commit aman → hapus file lama async
	if oldImageIDToDelete != nil {
		go func(imageID string) {
			if err := helpers.DeleteLocalFileImmediate(imageID); err != nil {
				log.Printf("Warning: Failed to delete old image file %s: %v", imageID, err)
			}
		}(oldImageIDToDelete.String())
	}

	// fetch terbaru (opsional)
	latest, err := s.ItemRepository.FindById(nil, updated.ID.String(), false)
	if err != nil {
		log.Printf("Warning: Item updated but failed to fetch updated data: %v", err)
		return updated, nil
	}
	return latest, nil
}

func (s *ItemService) DeleteItems(in *models.ItemIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	_ = ctx; _ = userInfo

	if len(in.IDs) == 0 {
		return errors.New("itemIds cannot be empty")
	}

	for _, id := range in.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for item %v: %v\n", id, tx.Error)
			return errors.New("error beginning transaction")
		}

		it, err := s.ItemRepository.FindById(tx, id.String(), true)
		if err != nil {
			tx.Rollback()
			if err == repositories.ErrItemNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("Item not found: %v\n", id)
				continue
			}
			log.Printf("Error finding item %v: %v\n", id, err)
			return errors.New("error finding item")
		}

		isHard := in.IsHardDelete == "hardDelete"
		if err := s.ItemRepository.Delete(tx, id.String(), isHard); err != nil {
			tx.Rollback()
			log.Printf("Error deleting item %v: %v\n", id, err)
			return errors.New("error deleting item")
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("Error committing delete for item %v: %v\n", id, err)
			return errors.New("error committing delete")
		}

		if isHard && it.ImageID != nil && it.ImageID.String() != "" {
			go func(imageID string) {
				if err := helpers.DeleteLocalFileImmediate(imageID); err != nil {
					log.Printf("Warning: Failed to delete image file %s: %v", imageID, err)
				}
			}(it.ImageID.String())
		}
	}
	return nil
}

func (s *ItemService) RestoreItems(in *models.ItemRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.Item, error) {
	_ = ctx; _ = userInfo

	if len(in.IDs) == 0 {
		return nil, errors.New("itemIds cannot be empty")
	}

	var restored []models.Item
	for _, id := range in.IDs {
		tx := configs.DB.Begin()
		if tx.Error != nil {
			log.Printf("Failed to begin transaction for item restore %v: %v\n", id, tx.Error)
			return nil, errors.New("error beginning transaction")
		}

		res, err := s.ItemRepository.Restore(tx, id.String())
		if err != nil {
			tx.Rollback()
			if err == repositories.ErrItemNotFound || errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("Item not found for restore: %v\n", id)
				continue
			}
			log.Printf("Error restoring item %v: %v\n", id, err)
			return nil, errors.New("error restoring item")
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("Error committing item restore %v: %v\n", id, err)
			return nil, errors.New("error committing item restore")
		}

		restoredItem, ferr := s.ItemRepository.FindById(nil, res.ID.String(), true)
		if ferr != nil {
			log.Printf("Error fetching restored item %v: %v\n", id, ferr)
			continue
		}
		restored = append(restored, *restoredItem)
	}
	return restored, nil
}