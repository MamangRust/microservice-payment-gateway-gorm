package tests

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// dbClusterContract is the authoritative service → bounded-context mapping.
// Every DB-backed service must pass the listed pkg/database constant to
// server.Config.DBCluster.
//
// This guards a real regression: the services once still passed the retired
// opaque identifiers "DB_A".."DB_D" while .env / docker.env / the Kubernetes
// app-config had already moved to DB_IDENTITY / DB_PAYMENT / DB_FINANCIAL.
// Because the old prefixes no longer exist, host/port/dbname resolved empty.
var dbClusterContract = map[string]string{
	"auth":        "IdentityCluster",
	"role":        "IdentityCluster",
	"user":        "IdentityCluster",
	"card":        "PaymentCluster",
	"merchant":    "PaymentCluster",
	"saldo":       "PaymentCluster",
	"topup":       "FinancialCluster",
	"transaction": "FinancialCluster",
	"transfer":    "FinancialCluster",
	"withdraw":    "FinancialCluster",
}

// clusterPrefixes are the env prefixes the contract above must resolve to.
var clusterPrefixes = []string{"DB_IDENTITY", "DB_PAYMENT", "DB_FINANCIAL"}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "justfile")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not locate repo root (justfile) above %s", dir)
		}
		dir = parent
	}
}

// dbClusterInFile parses a cmd/main.go and returns the qualified value assigned
// to the DBCluster field (e.g. "database.IdentityCluster").
func dbClusterInFile(path string) (string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return "", err
	}

	var got string
	ast.Inspect(file, func(n ast.Node) bool {
		kv, ok := n.(*ast.KeyValueExpr)
		if !ok {
			return true
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != "DBCluster" {
			return true
		}
		if sel, ok := kv.Value.(*ast.SelectorExpr); ok {
			if pkg, ok := sel.X.(*ast.Ident); ok {
				got = pkg.Name + "." + sel.Sel.Name
			}
		} else if lit, ok := kv.Value.(*ast.BasicLit); ok {
			// A string literal here means the service regressed to a raw
			// prefix (e.g. the retired "DB_A") instead of the names.go constant.
			got = lit.Value
		}
		return false
	})

	if got == "" {
		return "", fmt.Errorf("no DBCluster value found in %s", path)
	}
	return got, nil
}

// TestDBClusterContract pins each service to its bounded-context cluster.
func TestDBClusterContract(t *testing.T) {
	root := repoRoot(t)

	for svc, wantConst := range dbClusterContract {
		path := filepath.Join(root, "service", svc, "cmd", "main.go")
		got, err := dbClusterInFile(path)
		if err != nil {
			t.Errorf("service %s: %v", svc, err)
			continue
		}
		want := "database." + wantConst
		if got != want {
			t.Errorf("service %s: DBCluster = %q, want %q (see pkg/database/names.go)", svc, got, want)
		}
	}
}

// TestDBClusterPrefixesDefinedInEnv ensures the prefixes the services resolve
// actually exist in both env files, and that the base DB_HOST/DB_PORT/DB_NAME
// keys stay absent (their presence would silently mask a broken prefix).
func TestDBClusterPrefixesDefinedInEnv(t *testing.T) {
	root := repoRoot(t)
	envFiles := []string{
		filepath.Join(root, ".env"),
		filepath.Join(root, "deployments", "local", "docker.env"),
	}

	for _, envPath := range envFiles {
		data, err := os.ReadFile(envPath)
		if err != nil {
			t.Errorf("read %s: %v", envPath, err)
			continue
		}
		env := string(data)

		for _, prefix := range clusterPrefixes {
			for _, suffix := range []string{"_HOST", "_PORT", "_NAME"} {
				key := prefix + suffix
				if !strings.Contains(env, "\n"+key+"=") {
					t.Errorf("%s: missing required key %s", envPath, key)
				}
			}
		}

		for _, key := range []string{"DB_HOST", "DB_PORT", "DB_NAME"} {
			if strings.Contains(env, "\n"+key+"=") {
				t.Errorf("%s: base key %s must not be defined (services resolve per-context prefixes)", envPath, key)
			}
		}
	}
}
