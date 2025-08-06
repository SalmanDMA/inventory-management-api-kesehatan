package services

// type FileService struct {
// 	UploadRepository repositories.UploadRepository
// }

// func NewFileService(uploadRepository repositories.UploadRepository) *FileService {
// 	return &FileService{UploadRepository: uploadRepository}
// }

// func (service *FileService) GetFileByID(fileId string) (*models.ResponseGetFile, error) {
// 	file, err := service.UploadRepository.FindById(fileId, true)
// 	if err != nil {
// 		return nil, err
// 	}

// 	url := helpers.GetFileURL(file)

// 	return &models.ResponseGetFile{
// 		URL: url,
// 	}, nil
// }