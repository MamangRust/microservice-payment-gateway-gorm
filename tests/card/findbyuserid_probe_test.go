package card_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TestProbeFindByUserIdCard spawns the card service binary itself (immune to
// the sandbox's background-process reaping) and prints the raw
// FindByUserIdCard response for users 1..3.
func TestProbeFindByUserIdCard(t *testing.T) {
	bin := "/tmp/e2e-bin/card"
	if _, err := os.Stat(bin); err != nil {
		t.Skipf("card binary not found at %s (build via scripts/e2e-local.sh): %v", bin, err)
	}

	env := append(os.Environ(),
		"APP_ENV=test", "DB_DRIVER=postgres", "DB_HOST=localhost", "DB_PORT=5436",
		"DB_NAME=card_db", "DB_USERNAME=DRAGON", "DB_PASSWORD=DRAGON", "SECRET_KEY=yantopedia",
		"KAFKA_BROKERS=localhost:9092", "REDIS_ADDRS=localhost:6379", "REDIS_PASSWORD=dragon_knight",
		"REDIS_DB=0", "GRPC_USER_ADDR=localhost:50055",
	)
	// card's MigrationPath is "./migrations" (relative), so run it from its own dir.
	svcDir := filepath.Join("..", "..", "service", "card")
	cmd := exec.Command(bin)
	cmd.Dir = svcDir
	cmd.Env = env
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("start card: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	}()

	// wait for the gRPC port
	deadline := time.Now().Add(60 * time.Second)
	for {
		conn, err := net.DialTimeout("tcp", "127.0.0.1:50053", time.Second)
		if err == nil {
			conn.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("card service never came up: %v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	time.Sleep(500 * time.Millisecond)

	conn, err := grpc.NewClient("127.0.0.1:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewCardQueryServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, uid := range []int32{1, 2, 3} {
		resp, err := client.FindByUserIdCard(ctx, &pb.FindByUserIdCardRequest{UserId: uid})
		if err != nil {
			t.Logf("FindByUserIdCard(%d) error: %v", uid, err)
			continue
		}
		d := resp.GetData()
		if d == nil {
			t.Logf("FindByUserIdCard(%d): Data NIL; status=%q msg=%q", uid, resp.GetStatus(), resp.GetMessage())
			continue
		}
		t.Logf("FindByUserIdCard(%d) => id=%d user_id=%d card_number=%q type=%q cvv=%q provider=%q expire=%q",
			uid, d.GetId(), d.GetUserId(), d.GetCardNumber(), d.GetCardType(), d.GetCvv(), d.GetCardProvider(), d.GetExpireDate())
	}

	_ = fmt.Sprint()
}
