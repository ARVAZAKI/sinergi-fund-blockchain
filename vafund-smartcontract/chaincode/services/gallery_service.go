package services

import (
	"encoding/json"
	"fmt"
	"time"

	"vafund-chaincode/entities"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// GalleryService handles gallery-related operations
type GalleryService struct{}

// NewGalleryService creates a new gallery service instance
func NewGalleryService() *GalleryService {
	return &GalleryService{}
}

// Create creates a new gallery entry
func (gs *GalleryService) Create(ctx contractapi.TransactionContextInterface, id string, eventCode string, imageURL string, description string) error {
	exists, err := gs.Exists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("gallery %s already exists", id)
	}

	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("failed to get transaction timestamp: %v", err)
	}

	gallery := entities.NewGallery(
		id,
		eventCode,
		imageURL,
		description,
		time.Unix(timestamp.Seconds, int64(timestamp.Nanos)),
		ctx.GetStub().GetTxID(),
	)

	if !gallery.IsValid() {
		return fmt.Errorf("invalid gallery data")
	}

	galleryJSON, err := json.Marshal(gallery)
	if err != nil {
		return fmt.Errorf("failed to marshal gallery: %v", err)
	}

	return ctx.GetStub().PutState("GALLERY_"+id, galleryJSON)
}

// Update updates an existing gallery entry
func (gs *GalleryService) Update(ctx contractapi.TransactionContextInterface, id string, eventCode string, imageURL string, description string) error {
	exists, err := gs.Exists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("gallery %s does not exist", id)
	}

	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("failed to get transaction timestamp: %v", err)
	}

	gallery := entities.NewGallery(
		id,
		eventCode,
		imageURL,
		description,
		time.Unix(timestamp.Seconds, int64(timestamp.Nanos)),
		ctx.GetStub().GetTxID(),
	)

	if !gallery.IsValid() {
		return fmt.Errorf("invalid gallery data")
	}

	galleryJSON, err := json.Marshal(gallery)
	if err != nil {
		return fmt.Errorf("failed to marshal gallery: %v", err)
	}

	return ctx.GetStub().PutState("GALLERY_"+id, galleryJSON)
}

// Delete deletes a gallery entry
func (gs *GalleryService) Delete(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := gs.Exists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("gallery %s does not exist", id)
	}

	return ctx.GetStub().DelState("GALLERY_" + id)
}

// GetByID retrieves a gallery by ID
func (gs *GalleryService) GetByID(ctx contractapi.TransactionContextInterface, id string) (*entities.Gallery, error) {
	galleryJSON, err := ctx.GetStub().GetState("GALLERY_" + id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if galleryJSON == nil {
		return nil, fmt.Errorf("gallery %s does not exist", id)
	}

	var gallery entities.Gallery
	err = json.Unmarshal(galleryJSON, &gallery)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal gallery: %v", err)
	}

	return &gallery, nil
}

// GetAll retrieves all galleries
func (gs *GalleryService) GetAll(ctx contractapi.TransactionContextInterface) ([]*entities.Gallery, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("GALLERY_", "GALLERY_\uffff")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var galleries []*entities.Gallery
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var gallery entities.Gallery
		err = json.Unmarshal(queryResponse.Value, &gallery)
		if err != nil {
			continue // Skip invalid entries
		}

		if gallery.IsValid() {
			galleries = append(galleries, &gallery)
		}
	}

	return galleries, nil
}

// GetByEventCode retrieves galleries by event code
func (gs *GalleryService) GetByEventCode(ctx contractapi.TransactionContextInterface, eventCode string) ([]*entities.Gallery, error) {
	allGalleries, err := gs.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var eventGalleries []*entities.Gallery
	for _, gallery := range allGalleries {
		if gallery.EventCode == eventCode {
			eventGalleries = append(eventGalleries, gallery)
		}
	}

	return eventGalleries, nil
}

// Exists checks if a gallery with the given ID exists
func (gs *GalleryService) Exists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	galleryJSON, err := ctx.GetStub().GetState("GALLERY_" + id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return galleryJSON != nil, nil
}
