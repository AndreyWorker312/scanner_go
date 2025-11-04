package service

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"test_tcp/internal/config"
	"test_tcp/pkg/logger"
	"test_tcp/pkg/queue"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service struct {
	cfg *config.Config
	log logger.Logger
	mq  *queue.RabbitMQ
	mdb *mongo.Database
	mio *minio.Client
}

func Run(ctx context.Context, cfg *config.Config, log logger.Logger) error {
	// RabbitMQ
	mq, err := queue.New(cfg.RabbitMQURL, cfg.ScannerName)
	if err != nil {
		return fmt.Errorf("rabbitmq connect: %w", err)
	}
	defer mq.Close()
	log.Infof("Connected RabbitMQ, queue=%s", cfg.ScannerName)

	// MongoDB
	mongoCli, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return fmt.Errorf("mongo connect: %w", err)
	}
	defer mongoCli.Disconnect(ctx)
	mdb := mongoCli.Database(cfg.MongoDB)
	log.Infof("Connected MongoDB db=%s", cfg.MongoDB)

	// MinIO
	mio, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccess, cfg.MinIOSecret, ""),
		Secure: false,
	})
	if err != nil {
		return fmt.Errorf("minio connect: %w", err)
	}
	// Ensure bucket
	exists, err := mio.BucketExists(ctx, cfg.MinIOBucket)
	if err != nil {
		return fmt.Errorf("minio bucket check: %w", err)
	}
	if !exists {
		if err := mio.MakeBucket(ctx, cfg.MinIOBucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("minio make bucket: %w", err)
		}
	}
	log.Infof("Connected MinIO bucket=%s", cfg.MinIOBucket)

	s := &Service{cfg: cfg, log: log, mq: mq, mdb: mdb, mio: mio}
	msgs, err := mq.Consume(ctx)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}
	log.Infof("Waiting for TCP tasks...")

	for {
		select {
		case <-ctx.Done():
			return nil
		case d, ok := <-msgs:
			if !ok {
				return fmt.Errorf("rabbit channel closed")
			}
			// Handle message
			s.handle(ctx, d)
		}
	}
}

func (s *Service) handle(ctx context.Context, d queue.Delivery) {
	req, err := queue.ParseTCPRequest(d.Body)
	if err != nil {
		s.log.Errorf("bad msg: %v", err)
		s.mq.Nack(d)
		return
	}

	s.log.Infof("TCP read: %s:%s (task=%s)", req.Host, req.Port, req.TaskID)
	hexData, decoded, readErr := s.readTCP(ctx, req.Host, req.Port)
	var objKey string
	if readErr == nil {
		// store hex to MinIO
		objKey = fmt.Sprintf("%s_%d.hex", req.TaskID, time.Now().UnixNano())
		reader := strings.NewReader(bytesToHexLine(hexData))
		_, err := s.mio.PutObject(ctx, s.cfg.MinIOBucket, objKey, reader, int64(reader.Len()), minio.PutObjectOptions{ContentType: "text/plain"})
		if err != nil {
			s.log.Errorf("minio put: %v", err)
		}
		// store decoded to MongoDB
		coll := s.mdb.Collection(s.cfg.MongoColl)
		_, err = coll.InsertOne(ctx, bson.M{
			"task_id":        req.TaskID,
			"host":           req.Host,
			"port":           req.Port,
			"hex_object_key": objKey,
			"decoded_text":   decoded,
			"created_at":     time.Now(),
		})
		if err != nil {
			s.log.Errorf("mongo insert: %v", err)
		}
	}

	// reply if RPC
	if d.ReplyTo != "" {
		resp := queue.TCPResponse{TaskID: req.TaskID, Host: req.Host, Port: req.Port, HexObjectKey: objKey, DecodedText: decoded, Status: "completed"}
		if readErr != nil {
			resp.Status = "failed"
			resp.Error = readErr.Error()
		}
		if err := s.mq.Reply(d.ReplyTo, d.CorrelationId, resp); err != nil {
			s.log.Errorf("reply: %v", err)
		}
	}
	s.mq.Ack(d)
}

func (s *Service) readTCP(ctx context.Context, host, port string) (raw []byte, decoded string, err error) {
	addr := net.JoinHostPort(host, port)
	conn, err := (&net.Dialer{Timeout: s.cfg.ConnTimeout}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, "", err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(s.cfg.ReadTimeout))

	buf := make([]byte, 0, 8192)
	tmp := make([]byte, 4096)
	for {
		n, er := conn.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if er != nil {
			break
		}
	}
	return buf, humanString(buf), nil
}

func bytesToHexLine(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, v := range b {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(fmt.Sprintf("%02X", v))
	}
	return sb.String()
}

func humanString(b []byte) string {
	out := make([]byte, 0, len(b))
	for _, x := range b {
		switch x {
		case '\r':
			out = append(out, '\\', 'r')
		case '\n':
			out = append(out, '\\', 'n')
		case '\t':
			out = append(out, '\\', 't')
		default:
			if x >= 32 && x <= 126 {
				out = append(out, x)
			}
		}
	}
	return string(out)
}
