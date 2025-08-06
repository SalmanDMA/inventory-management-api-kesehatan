package services

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ItemService struct {
	UploadRepository repositories.UploadRepository
	ItemRepository repositories.ItemRepository
	ItemHistoryRepository repositories.ItemHistoryRepository
}

func NewItemService(uploadRepo repositories.UploadRepository, itemRepo repositories.ItemRepository, itemHistoryRepo repositories.ItemHistoryRepository) *ItemService {
	return &ItemService{
		UploadRepository: uploadRepo,
		ItemRepository: itemRepo,
		ItemHistoryRepository: itemHistoryRepo,
	}
}

func (service *ItemService) GetAllItems() ([]models.ResponseGetItem, error) {
	items, err := service.ItemRepository.FindAll()
	if err != nil {
		return nil, err
	}

	var itemsResponse []models.ResponseGetItem
	for _, item := range items {
		itemsResponse = append(itemsResponse, models.ResponseGetItem{
			ID:          item.ID,
			Name:        item.Name,
			Code:        item.Code,
			Price:       item.Price,
			ImageID:     item.ImageID,
			Image:       item.Image,
			CategoryID:  item.CategoryID,
			Category:    item.Category,
			Description: item.Description,
			Stock:  item.Stock,
			ItemHistories: item.ItemHistories,
			CreatedAt: 	item.CreatedAt,
			UpdatedAt: 	item.UpdatedAt,
			DeletedAt: 	item.DeletedAt,
		})
	}

	return itemsResponse, nil
}

func (service *ItemService) GetAllItemsPaginated(req *models.PaginationRequest, userInfo *models.User) (*models.ItemPaginatedResponse, error) {
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

	items, totalCount, err := service.ItemRepository.FindAllPaginated(req)
	if err != nil {
		return nil, err
	}

	itemsResponse := []models.ResponseGetItem{}
	for _, item := range items {
		itemsResponse = append(itemsResponse, models.ResponseGetItem{
			ID:          item.ID,
			Name:        item.Name,
			Code:        item.Code,
			Price:       item.Price,
			ImageID:     item.ImageID,
			Image:       item.Image,
			CategoryID:  item.CategoryID,
			Category:    item.Category,
			Description: item.Description,
			Stock:  item.Stock,
			ItemHistories: item.ItemHistories,
			CreatedAt: 	item.CreatedAt,
			UpdatedAt: 	item.UpdatedAt,
			DeletedAt: 	item.DeletedAt,
		})
	}

	totalPages := int((totalCount + int64(req.Limit) - 1) / int64(req.Limit))
	hasNext := req.Page < totalPages
	hasPrev := req.Page > 1

	paginationResponse := models.PaginationResponse{
		CurrentPage:  req.Page,
		PerPage:      req.Limit,
		TotalPages:   totalPages,
		TotalRecords: totalCount,
		HasNext:      hasNext,
		HasPrev:      hasPrev,
	}

	return &models.ItemPaginatedResponse{
		Data:       itemsResponse,
		Pagination: paginationResponse,
	}, nil
}

func (service *ItemService) GetItemByID(itemId string) (*models.ResponseGetItem, error) {
	item, err := service.ItemRepository.FindById(itemId, false)

	if err != nil {
		return nil, err
	}

	return &models.ResponseGetItem{
		 ID:          item.ID,
			Name:        item.Name,
			Code:        item.Code,
			Price:       item.Price,
			ImageID:     item.ImageID,
			Image:       item.Image,
			CategoryID:  item.CategoryID,
			Category:    item.Category,
			Description: item.Description,
			Stock:  					item.Stock,
			ItemHistories: item.ItemHistories,
	}, nil
}

func (service *ItemService) CreateItem(itemRequest *models.ItemCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Item, error) {
	if _, err := service.ItemRepository.FindByName(itemRequest.Name); err == nil {
		return nil, errors.New("item already exists") 
	} else if err != repositories.ErrItemNotFound {
		 return nil, errors.New("error checking item: " + err.Error())
	}

	newItem := &models.Item{
		ID:          uuid.New(),
		Name:        itemRequest.Name,
		Code:        itemRequest.Code,
		Price:       itemRequest.Price,
		Stock:       itemRequest.Stock,
		CategoryID:  itemRequest.CategoryID,
		Description: itemRequest.Description,
	}

	file, err := ctx.FormFile("image")
	if err == nil {
		uuidStr, err := helpers.SaveFile(ctx, file, "items")
		if err != nil {
			return nil, err
		}
		imageUUID, err := uuid.Parse(uuidStr)
		if err != nil {
			return nil, err
		}
		newItem.ImageID = &imageUUID
	}

	item, err := service.ItemRepository.Insert(newItem)
	if err != nil {
		return nil, err
	}

	if _, err := service.ItemHistoryRepository.Insert(&models.ItemHistory{
		ID:          uuid.New(),
		ItemID:      item.ID,
		ChangeType:  "create_price",
		OldPrice:    0,
		NewPrice:    item.Price,
		CurrentPrice: item.Price,
		CreatedBy:   &userInfo.ID,
		UpdatedBy:   &userInfo.ID,
		Description: "Initial price set to " + strconv.Itoa(item.Price),
	}); err != nil {
		return nil, err
	}

	if _, err := service.ItemHistoryRepository.Insert(&models.ItemHistory{
		ID:          uuid.New(),
		ItemID:      item.ID,
		ChangeType:  "create_stock",
		OldStock:    0,
		NewStock:    item.Stock,
		CurrentStock: item.Stock,
		CreatedBy:   &userInfo.ID,
		UpdatedBy:   &userInfo.ID,
		Description: "Initial stock set to " + strconv.Itoa(item.Stock),
	}); err != nil {
		return nil, err
	}

	return item, nil
}

func (service *ItemService) UpdateItem(itemID string, itemUpdate *models.ItemUpdateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Item, error) {
	itemExists, err := service.ItemRepository.FindById(itemID, false)
	if err != nil {
		return nil, err
	}
	if itemExists == nil {
		return nil, errors.New("item not found")
	}

	if itemUpdate.Name != "" {
		itemExists.Name = itemUpdate.Name
	}
	if itemUpdate.Code != "" {
		itemExists.Code = itemUpdate.Code
	}
	if itemUpdate.CategoryID != uuid.Nil {
		itemExists.CategoryID = itemUpdate.CategoryID
	}
	if itemUpdate.Description != "" {
		itemExists.Description = itemUpdate.Description
	}

	file, err := ctx.FormFile("image")
	if err == nil && file != nil {
		imageIDStr, err := helpers.SaveFile(ctx, file, "items")
		if err != nil {
			return nil, fmt.Errorf("failed to save image: %w", err)
		}

		newImageID, err := uuid.Parse(imageIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid image ID: %w", err)
		}

		oldImageID := itemExists.ImageID
		itemExists.ImageID = &newImageID

		updatedUser, err := service.ItemRepository.Update(itemExists)
		if err != nil {
			return nil, fmt.Errorf("error updating itemExists with image: %w", err)
		}

		if oldImageID != nil {
			service.UploadRepository.Delete(oldImageID.String(), false)
		}

		return updatedUser, nil
	}

	updateItem, err := service.ItemRepository.Update(itemExists)
	if err != nil {
		return nil, err
	}

	return updateItem, nil
}

func (service *ItemService) DeleteItems(itemRequest *models.ItemIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	for _, itemId := range itemRequest.IDs {
		item, err := service.ItemRepository.FindById(itemId.String(), false)
		if err != nil {
			if err == repositories.ErrItemNotFound {
				log.Printf("Item not found: %v\n", itemId)
				continue
			}
			log.Printf("Error finding item %v: %v\n", itemId, err)
			return errors.New("error finding useitemr")
		}

		if itemRequest.IsHardDelete == "hardDelete" {
			if item.ImageID != nil && item.ImageID.String() != "" {
				if err := service.UploadRepository.Delete(item.ImageID.String(), false); err != nil {
					log.Printf("Error deleting image file for item %v: %v\n", itemId, err)
					return errors.New("error deleting image file")
				}
			}

			if err := service.ItemRepository.Delete(itemId.String(), true); err != nil {
				log.Printf("Error hard deleting item %v: %v\n", itemId, err)
				return errors.New("error hard deleting item")
			}
		} else {
			if err := service.ItemRepository.Delete(itemId.String(), false); err != nil {
				log.Printf("Error soft deleting item %v: %v\n", itemId, err)
				return errors.New("error soft deleting item")
			}
		}
	}

	return nil
}

func (service *ItemService) RestoreItems(item *models.ItemRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.Item, error) {
	var restoredItems []models.Item

	for _, itemId := range item.IDs {
		item := &models.Item{ID: itemId}

		restoredItem, err := service.ItemRepository.Restore(item, itemId.String())
		if err != nil {
			if err == repositories.ErrItemNotFound {
				log.Printf("Item not found: %v\n", itemId)
				continue
			}
			log.Printf("Error restoring item %v: %v\n", itemId, err)
			return nil, errors.New("error restoring item")
		}

		restoredItems = append(restoredItems, *restoredItem)
	}

	return restoredItems, nil
}