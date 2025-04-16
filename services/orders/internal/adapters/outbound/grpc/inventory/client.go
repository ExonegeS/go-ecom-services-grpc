package grpc

import (
	"context"

	"github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/domain/entity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type InventoryClient struct {
	conn   *grpc.ClientConn
	client InventoryServiceClient
}

func NewInventoryClient(address string) (*InventoryClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &InventoryClient{
		conn:   conn,
		client: NewInventoryServiceClient(conn),
	}, nil
}

func (c *InventoryClient) GetProduct(ctx context.Context, productID entity.UUID) (*entity.OrderItem, error) {
	resp, err := c.client.GetProductByID(ctx, &GetProductRequest{Id: string(productID)})
	if err != nil {
		return nil, err
	}

	product := resp.GetProduct()

	entity := &entity.OrderItem{
		ProductID:    entity.UUID(product.GetId()),
		ProductName:  product.GetName(),
		ProductPrice: product.GetPrice(),
		Quantity:     int64(product.GetQuantity()),
		CreatedAt:    product.CreatedAt.AsTime(),
		UpdatedAt:    product.UpdatedAt.AsTime(),
	}

	return entity, nil
}

func (c *InventoryClient) ReserveItem(ctx context.Context, itemID entity.UUID, quantity int64) (bool, error) {
	// TODO: Make it to be a worker that collects requests and sends a combined request once in a second, instead of constant updating
	_, err := c.client.ReserveProducts(ctx, &ReserveProductRequest{
		Id:       itemID.String(),
		Quantity: int32(quantity),
	})

	return true, err
}
