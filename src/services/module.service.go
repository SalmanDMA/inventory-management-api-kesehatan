package services

import (
	"errors"
	"fmt"

	"github.com/SalmanDMA/inventory-app/backend/src/configs"
	"github.com/SalmanDMA/inventory-app/backend/src/helpers"
	"github.com/SalmanDMA/inventory-app/backend/src/models"
	"github.com/SalmanDMA/inventory-app/backend/src/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ModuleService struct {
	ModuleRepository repositories.ModuleRepository
}

func NewModuleService(moduleRepository repositories.ModuleRepository) *ModuleService {
	return &ModuleService{
		ModuleRepository: moduleRepository,
	}
}

// ==============================
// Reads
// ==============================

func (s *ModuleService) GetAllModules() ([]models.ResponseGetModule, error) {
	// default: hanya data aktif (tanpa soft-deleted) jika repo kamu punya varian FindAllActive(nil)
	// sekarang tetap pakai FindAll(nil) sesuai implementasi yang ada
	modules, err := s.ModuleRepository.FindAll(nil)
	if err != nil {
		return nil, err
	}

	resp := make([]models.ResponseGetModule, 0, len(modules))
	for _, m := range modules {
		resp = append(resp, models.ResponseGetModule{
			ID:           m.ID,
			Name:         m.Name,
			ParentID:     m.ParentID,
			Parent:       m.Parent,
			ModuleTypeID: m.ModuleTypeID,
			ModuleType:   m.ModuleType,
			Route:        m.Route,
			Path:         m.Path,
			Icon:         m.Icon,
			Children:     m.Children,
			Description:  m.Description,
			RoleModules:  m.RoleModules,
			DeletedAt:    m.DeletedAt,
		})
	}
	return resp, nil
}

// ==============================
// Mutations (transaction-aware via configs.DB.Begin())
// ==============================

func (s *ModuleService) CreateModule(req *models.ModuleCreateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Module, error) {
	_ = ctx; _ = userInfo

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("begin tx: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	parentID, err := s.resolveParentID(tx, req.ParentID)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	route := req.Route
	if route != "" {
		route, err = helpers.FormatRoute(route)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}

	newModule := &models.Module{
		Name:         helpers.CapitalizeTitle(req.Name),
		ParentID:     parentID,
		ModuleTypeID: req.ModuleTypeID,
		Route:        route,
		Path:         req.Path,
		Icon:         req.Icon,
		Description:  req.Description,
	}

	created, err := s.ModuleRepository.Insert(tx, newModule)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return created, nil
}

func (s *ModuleService) UpdateModule(moduleId int, req *models.ModuleUpdateRequest, ctx *fiber.Ctx, userInfo *models.User) (*models.Module, error) {
	_ = ctx; _ = userInfo

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("begin tx: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	m, err := s.ModuleRepository.FindById(tx, moduleId, true)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	// Parent handling
	if req.ParentID != nil {
		parentID, err := s.resolveParentID(tx, req.ParentID)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		// Cegah siklus: parent tidak boleh menjadi descendant dari m
		if parentID != nil {
			isDescendant, err := s.isDescendant(tx, *parentID, m.ID)
			if err != nil {
				_ = tx.Rollback()
				return nil, err
			}
			if isDescendant || *parentID == m.ID {
				_ = tx.Rollback()
				return nil, fmt.Errorf("invalid parent: cycle detected")
			}
		}
		m.ParentID = parentID
		m.Parent = nil
	}

	// ModuleType
	if req.ModuleTypeID != uuid.Nil {
		m.ModuleTypeID = req.ModuleTypeID
	}

	// Route
	if req.Route != "" {
		route, err := helpers.FormatRoute(req.Route)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		m.Route = route
	}

	if req.Name != "" {
		m.Name = helpers.CapitalizeTitle(req.Name)
	}
	if req.Path != "" {
		m.Path = req.Path
	}
	if req.Icon != "" {
		m.Icon = req.Icon
	}
	if req.Description != "" {
		m.Description = req.Description
	}

	updated, err := s.ModuleRepository.Update(tx, m)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return updated, nil
}

func (s *ModuleService) DeleteModule(req *models.ModuleIsHardDeleteRequest, ctx *fiber.Ctx, userInfo *models.User) error {
	_ = ctx; _ = userInfo

	if len(req.IDs) == 0 {
		return fmt.Errorf("moduleIds cannot be empty")
	}
	isHard := req.IsHardDelete == "hardDelete"

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin tx: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	// Ambil semua module (Unscoped untuk bisa juga detach anak dari parent soft-deleted)
	all, err := s.ModuleRepository.FindAll(tx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("get all modules: %w", err)
	}

	toDelete := make(map[int]bool, len(req.IDs))
	for _, id := range req.IDs {
		toDelete[id] = true
	}

	// Build adjacency list supaya collect O(n)
	childrenMap := buildChildrenMap(all)
	collectDescendants(toDelete, childrenMap)

	// Detach anak yg parent-nya akan dihapus, tapi parent-nya tidak ikut dihapus
	for i := range all {
		m := &all[i]
		if m.ParentID != nil && toDelete[m.ID] && !toDelete[*m.ParentID] {
			m.ParentID = nil
			if _, err := s.ModuleRepository.Update(tx, m); err != nil {
				_ = tx.Rollback()
				return err
			}
		}
	}

	// Hapus
	for id := range toDelete {
		if _, err := s.ModuleRepository.FindById(tx, id, false); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := s.ModuleRepository.Delete(tx, id, isHard); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (s *ModuleService) RestoreModule(req *models.ModuleRestoreRequest, ctx *fiber.Ctx, userInfo *models.User) ([]models.Module, error) {
	_ = ctx; _ = userInfo

	if len(req.IDs) == 0 {
		return nil, fmt.Errorf("moduleIds cannot be empty")
	}

	tx := configs.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("begin tx: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	all, err := s.ModuleRepository.FindAll(tx)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("get all modules: %w", err)
	}

	toRestore := make(map[int]bool, len(req.IDs))
	for _, id := range req.IDs {
		toRestore[id] = true
	}

	childrenMap := buildChildrenMap(all)
	collectDescendants(toRestore, childrenMap)

	for i := range all {
		m := &all[i]
		if m.ParentID != nil && toRestore[m.ID] && !toRestore[*m.ParentID] {
			m.ParentID = nil
			if _, err := s.ModuleRepository.Update(tx, m); err != nil {
				_ = tx.Rollback()
				return nil, err
			}
		}
	}

	var restored []models.Module
	for id := range toRestore {
		got, err := s.ModuleRepository.Restore(tx, id)
		if err != nil {
			if errors.Is(err, repositories.ErrModuleNotFound) {
				continue
			}
			_ = tx.Rollback()
			return nil, err
		}
		restored = append(restored, *got)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return restored, nil
}

// ==============================
// Helpers
// ==============================

func (s *ModuleService) resolveParentID(tx *gorm.DB, reqParentID *int) (*int, error) {
	if reqParentID == nil {
		return nil, nil
	}
	parent, err := s.ModuleRepository.FindById(tx, *reqParentID, false)
	if err != nil {
		return nil, err
	}
	if parent != nil && parent.Name != "Root" {
		id := parent.ID
		return &id, nil
	}
	return nil, nil
}

func (s *ModuleService) isDescendant(tx *gorm.DB, candidateParent int, nodeId int) (bool, error) {
	all, err := s.ModuleRepository.FindAll(tx)
	if err != nil {
		return false, err
	}
	childrenMap := buildChildrenMap(all)
	stack := []int{nodeId}
	seen := map[int]bool{nodeId: true}
	for len(stack) > 0 {
		cur := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, ch := range childrenMap[cur] {
			if ch == candidateParent {
				return true, nil
			}
			if !seen[ch] {
				seen[ch] = true
				stack = append(stack, ch)
			}
		}
	}
	return false, nil
}

func buildChildrenMap(all []models.Module) map[int][]int {
	mp := make(map[int][]int, len(all))
	for _, m := range all {
		if m.ParentID != nil {
			mp[*m.ParentID] = append(mp[*m.ParentID], m.ID)
		}
	}
	return mp
}

func collectDescendants(target map[int]bool, childrenMap map[int][]int) {
	stack := make([]int, 0, len(target))
	for id := range target {
		stack = append(stack, id)
	}
	for len(stack) > 0 {
		p := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, child := range childrenMap[p] {
			if !target[child] {
				target[child] = true
				stack = append(stack, child)
			}
		}
	}
}
