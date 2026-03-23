package rabbitmq

import (
	"context"
	"log"
	"time"

	"backend/domain/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository struct {
	db *Database
}

func NewRepository(db *Database) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SaveARPHistory(record *models.ARPHistoryRecord) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	record.CreatedAt = time.Now()
	_, err := r.db.ARPCollection().InsertOne(ctx, record)
	if err != nil {
		log.Printf("Error saving ARP history: %v", err)
		return err
	}

	log.Printf("ARP history saved successfully for task: %s", record.TaskID)
	return nil
}

func (r *Repository) GetARPHistory(limit int) ([]models.ARPHistoryRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := r.db.ARPCollection().Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var records []models.ARPHistoryRecord
	if err = cursor.All(ctx, &records); err != nil {
		return nil, err
	}

	return records, nil
}

func (r *Repository) GetARPHistoryByIPRange(ipRange string, limit int) ([]models.ARPHistoryRecord, error) {
	if ipRange == "" {
		return r.GetARPHistory(limit)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"ip_range": ipRange}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	cursor, err := r.db.ARPCollection().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var records []models.ARPHistoryRecord
	if err = cursor.All(ctx, &records); err != nil {
		return nil, err
	}
	return records, nil
}

func (r *Repository) GetARPHistoryByID(id string) (*models.ARPHistoryRecord, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var rec models.ARPHistoryRecord
	err = r.db.ARPCollection().FindOne(ctx, bson.M{"_id": objID}).Decode(&rec)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *Repository) DeleteARPHistory() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := r.db.ARPCollection().DeleteMany(ctx, bson.D{})
	if err != nil {
		log.Printf("Error deleting ARP history: %v", err)
		return err
	}

	log.Printf("Deleted %d ARP history records", result.DeletedCount)
	return nil
}

func (r *Repository) SaveICMPHistory(record *models.ICMPHistoryRecord) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	record.CreatedAt = time.Now()
	_, err := r.db.ICMPCollection().InsertOne(ctx, record)
	if err != nil {
		log.Printf("Error saving ICMP history: %v", err)
		return err
	}

	log.Printf("ICMP history saved successfully for task: %s", record.TaskID)
	return nil
}

func (r *Repository) GetICMPHistory(limit int) ([]models.ICMPHistoryRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := r.db.ICMPCollection().Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var records []models.ICMPHistoryRecord
	if err = cursor.All(ctx, &records); err != nil {
		return nil, err
	}

	return records, nil
}

func (r *Repository) DeleteICMPHistory() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := r.db.ICMPCollection().DeleteMany(ctx, bson.D{})
	if err != nil {
		log.Printf("Error deleting ICMP history: %v", err)
		return err
	}

	log.Printf("Deleted %d ICMP history records", result.DeletedCount)
	return nil
}

func (r *Repository) GetICMPHistoryByTargets(targets []string, limit int) ([]models.ICMPHistoryRecord, error) {
	if len(targets) == 0 {
		return r.GetICMPHistory(limit)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"targets": bson.M{"$in": targets}}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	cursor, err := r.db.ICMPCollection().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var records []models.ICMPHistoryRecord
	if err = cursor.All(ctx, &records); err != nil {
		return nil, err
	}
	return records, nil
}

func (r *Repository) GetICMPHistoryByID(id string) (*models.ICMPHistoryRecord, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var rec models.ICMPHistoryRecord
	err = r.db.ICMPCollection().FindOne(ctx, bson.M{"_id": objID}).Decode(&rec)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *Repository) SaveNmapTcpUdpHistory(record *models.NmapTcpUdpHistoryRecord) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	record.CreatedAt = time.Now()
	_, err := r.db.NmapTcpUdpCollection().InsertOne(ctx, record)
	if err != nil {
		log.Printf("Error saving Nmap TCP/UDP history: %v", err)
		return err
	}

	log.Printf("Nmap TCP/UDP history saved successfully for task: %s", record.TaskID)
	return nil
}

func (r *Repository) GetNmapTcpUdpHistory(limit int) ([]models.NmapTcpUdpHistoryRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := r.db.NmapTcpUdpCollection().Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var records []models.NmapTcpUdpHistoryRecord
	if err = cursor.All(ctx, &records); err != nil {
		return nil, err
	}

	return records, nil
}

func (r *Repository) DeleteNmapTcpUdpHistory() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := r.db.NmapTcpUdpCollection().DeleteMany(ctx, bson.D{})
	if err != nil {
		log.Printf("Error deleting Nmap TCP/UDP history: %v", err)
		return err
	}

	log.Printf("Deleted %d Nmap TCP/UDP history records", result.DeletedCount)
	return nil
}

func (r *Repository) GetNmapTcpUdpHistoryByIP(ip string, limit int) ([]models.NmapTcpUdpHistoryRecord, error) {
	if ip == "" {
		return r.GetNmapTcpUdpHistory(limit)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"ip": ip}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	cursor, err := r.db.NmapTcpUdpCollection().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var records []models.NmapTcpUdpHistoryRecord
	if err = cursor.All(ctx, &records); err != nil {
		return nil, err
	}
	return records, nil
}

func (r *Repository) GetNmapTcpUdpHistoryByID(id string) (*models.NmapTcpUdpHistoryRecord, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var rec models.NmapTcpUdpHistoryRecord
	err = r.db.NmapTcpUdpCollection().FindOne(ctx, bson.M{"_id": objID}).Decode(&rec)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *Repository) SaveNmapOsDetectionHistory(record *models.NmapOsDetectionHistoryRecord) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	record.CreatedAt = time.Now()
	_, err := r.db.NmapOsDetectionCollection().InsertOne(ctx, record)
	if err != nil {
		log.Printf("Error saving Nmap OS Detection history: %v", err)
		return err
	}

	log.Printf("Nmap OS Detection history saved successfully for task: %s", record.TaskID)
	return nil
}

func (r *Repository) GetNmapOsDetectionHistory(limit int) ([]models.NmapOsDetectionHistoryRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := r.db.NmapOsDetectionCollection().Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var records []models.NmapOsDetectionHistoryRecord
	if err = cursor.All(ctx, &records); err != nil {
		return nil, err
	}

	return records, nil
}

func (r *Repository) DeleteNmapOsDetectionHistory() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := r.db.NmapOsDetectionCollection().DeleteMany(ctx, bson.D{})
	if err != nil {
		log.Printf("Error deleting Nmap OS Detection history: %v", err)
		return err
	}

	log.Printf("Deleted %d Nmap OS Detection history records", result.DeletedCount)
	return nil
}

func (r *Repository) GetNmapOsDetectionHistoryByIP(ip string, limit int) ([]models.NmapOsDetectionHistoryRecord, error) {
	if ip == "" {
		return r.GetNmapOsDetectionHistory(limit)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"ip": ip}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	cursor, err := r.db.NmapOsDetectionCollection().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var records []models.NmapOsDetectionHistoryRecord
	if err = cursor.All(ctx, &records); err != nil {
		return nil, err
	}
	return records, nil
}

func (r *Repository) GetNmapOsDetectionHistoryByID(id string) (*models.NmapOsDetectionHistoryRecord, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var rec models.NmapOsDetectionHistoryRecord
	err = r.db.NmapOsDetectionCollection().FindOne(ctx, bson.M{"_id": objID}).Decode(&rec)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *Repository) SaveNmapHostDiscoveryHistory(record *models.NmapHostDiscoveryHistoryRecord) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	record.CreatedAt = time.Now()
	_, err := r.db.NmapHostDiscoveryCollection().InsertOne(ctx, record)
	if err != nil {
		log.Printf("Error saving Nmap Host Discovery history: %v", err)
		return err
	}

	log.Printf("Nmap Host Discovery history saved successfully for task: %s", record.TaskID)
	return nil
}

func (r *Repository) GetNmapHostDiscoveryHistory(limit int) ([]models.NmapHostDiscoveryHistoryRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := r.db.NmapHostDiscoveryCollection().Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var records []models.NmapHostDiscoveryHistoryRecord
	if err = cursor.All(ctx, &records); err != nil {
		return nil, err
	}

	return records, nil
}

func (r *Repository) DeleteNmapHostDiscoveryHistory() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := r.db.NmapHostDiscoveryCollection().DeleteMany(ctx, bson.D{})
	if err != nil {
		log.Printf("Error deleting Nmap Host Discovery history: %v", err)
		return err
	}

	log.Printf("Deleted %d Nmap Host Discovery history records", result.DeletedCount)
	return nil
}

func (r *Repository) GetNmapHostDiscoveryHistoryByIP(ip string, limit int) ([]models.NmapHostDiscoveryHistoryRecord, error) {
	if ip == "" {
		return r.GetNmapHostDiscoveryHistory(limit)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"ip": ip}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	cursor, err := r.db.NmapHostDiscoveryCollection().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var records []models.NmapHostDiscoveryHistoryRecord
	if err = cursor.All(ctx, &records); err != nil {
		return nil, err
	}
	return records, nil
}

func (r *Repository) GetNmapHostDiscoveryHistoryByID(id string) (*models.NmapHostDiscoveryHistoryRecord, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var rec models.NmapHostDiscoveryHistoryRecord
	err = r.db.NmapHostDiscoveryCollection().FindOne(ctx, bson.M{"_id": objID}).Decode(&rec)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *Repository) SaveTCPHistory(record *models.TCPHistoryRecord) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	record.CreatedAt = time.Now()
	_, err := r.db.TCPCollection().InsertOne(ctx, record)
	if err != nil {
		log.Printf("Error saving TCP history: %v", err)
		return err
	}

	log.Printf("TCP history saved successfully for task: %s", record.TaskID)
	return nil
}

func (r *Repository) GetTCPHistory(limit int) ([]models.TCPHistoryRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := r.db.TCPCollection().Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var records []models.TCPHistoryRecord
	if err = cursor.All(ctx, &records); err != nil {
		return nil, err
	}

	return records, nil
}

func (r *Repository) GetTCPHistoryByHostPort(host, port string, limit int) ([]models.TCPHistoryRecord, error) {
	if host == "" && port == "" {
		return r.GetTCPHistory(limit)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{}
	if host != "" {
		filter["host"] = host
	}
	if port != "" {
		filter["port"] = port
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	cursor, err := r.db.TCPCollection().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var records []models.TCPHistoryRecord
	if err = cursor.All(ctx, &records); err != nil {
		return nil, err
	}
	return records, nil
}

func (r *Repository) GetTCPHistoryByID(id string) (*models.TCPHistoryRecord, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var rec models.TCPHistoryRecord
	err = r.db.TCPCollection().FindOne(ctx, bson.M{"_id": objID}).Decode(&rec)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *Repository) DeleteTCPHistory() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := r.db.TCPCollection().DeleteMany(ctx, bson.D{})
	if err != nil {
		log.Printf("Error deleting TCP history: %v", err)
		return err
	}

	log.Printf("Deleted %d TCP history records", result.DeletedCount)
	return nil
}
