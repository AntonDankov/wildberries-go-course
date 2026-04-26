package repository

import (
	"context"
	"fmt"
	"wildberries-go-course/L0/database"

	models "wildberries-go-course/L0/model"
)

type OrderRepository struct {
	db *database.Database
}

type OrderRepositoryInterface interface {
	GetOrderByID(ctx context.Context, orderUUID *string) (models.Order, error)
	InsertOrder(ctx context.Context, order *models.Order) error
}

func NewOrderRepository(db *database.Database) *OrderRepository {
	return &OrderRepository{db: db}
}

func (repo *OrderRepository) InsertOrder(ctx context.Context, order *models.Order) error {
	tx, err := repo.db.Pool.Begin(ctx)
	_, err = tx.Exec(ctx,
		`INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shard_key, sm_id, date_created, oof_shard)
                 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature, order.CustomerID,
		order.DeliveryService, order.Shardkey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO deliveries (order_uid, name, phone, zip, city, address, region, email)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return fmt.Errorf("failed to insert delivery: %w", err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO payments (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
                   VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		order.OrderUID, order.Payment.RequestID, order.Payment.Currency, order.Payment.Provider,
		order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost,
		order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		return fmt.Errorf("failed to insert payment: %w", err)
	}

	for _, item := range order.Items {
		_, err = tx.Exec(ctx,
			`INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
                             VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			return fmt.Errorf("failed to insert item: %w", err)
		}

	}
	tx.Commit(ctx)
	return nil
}

func (repo *OrderRepository) GetOrderByID(ctx context.Context, orderUUID *string) (models.Order, error) {
	order := models.Order{}
	query := `SELECT order_uid, track_number, entry, locale, internal_signature, 
            customer_id, delivery_service, shard_key, sm_id, date_created, oof_shard 
            FROM orders WHERE order_uid = $1`

	err := repo.db.Pool.QueryRow(ctx, query, orderUUID).Scan(
		&order.OrderUID,
		&order.TrackNumber,
		&order.Entry,
		&order.Locale,
		&order.InternalSignature,
		&order.CustomerID,
		&order.DeliveryService,
		&order.Shardkey,
		&order.SmID,
		&order.DateCreated,
		&order.OofShard,
	)
	if err != nil {
		return order, fmt.Errorf("failed to get order: %w", err)
	}

	query = `SELECT order_uid, name, phone, zip, city,
        address, region, email
        FROM   deliveries WHERE  order_uid = $1`

	err = repo.db.Pool.QueryRow(ctx, query, orderUUID).Scan(
		&order.Delivery.OrderUID, &order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip, &order.Delivery.City,
		&order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email,
	)
	if err != nil {
		return order, fmt.Errorf("failed to get order: %w", err)
	}

	query = `SELECT transaction, request_id, currency, provider,
                      amount, payment_dt, bank, delivery_cost,
                      goods_total, custom_fee
               FROM   payments
               WHERE  transaction = $1`

	err = repo.db.Pool.QueryRow(ctx, query, orderUUID).Scan(
		&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency, &order.Payment.Provider,
		&order.Payment.Amount, &order.Payment.PaymentDt, &order.Payment.Bank, &order.Payment.DeliveryCost,
		&order.Payment.GoodsTotal, &order.Payment.CustomFee,
	)
	if err != nil {
		return order, fmt.Errorf("failed to get payment: %w", err)
	}

	query = `SELECT order_uid, chrt_id, track_number, price, rid,
          name, sale, size, total_price, nm_id, brand, status
          FROM   items
          WHERE  order_uid = $1`

	rows, err := repo.db.Pool.Query(ctx, query, orderUUID)
	if err != nil {
		return order, fmt.Errorf("query items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.Item
		if err := rows.Scan(
			&item.OrderID, &item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid,
			&item.Name, &item.Sale, &item.Size, &item.TotalPrice,
			&item.NmID, &item.Brand, &item.Status,
		); err != nil {
			return order, fmt.Errorf("scan item row: %w", err)
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}

func (repo *OrderRepository) GetLastNOrders(ctx context.Context, n int) ([]models.Order, error) {
	var orders []models.Order

	query := `SELECT order_uid 
            FROM orders 
            ORDER BY date_created DESC LIMIT $1`

	rows, err := repo.db.Pool.Query(ctx, query, n)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}
	defer rows.Close()

	var orderUUID string

	for rows.Next() {
		if err := rows.Scan(
			&orderUUID,
		); err != nil {
			return nil, fmt.Errorf("scan order row: %w", err)
		}

		order, err := repo.GetOrderByID(ctx, &orderUUID)
		if err != nil {
			return nil, fmt.Errorf("load order details for %s: %w", order.OrderUID, err)
		}

		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return orders, nil
}
