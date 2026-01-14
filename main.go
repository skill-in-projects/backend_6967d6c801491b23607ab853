package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "os"
    "strings"
    "strconv"

    "backend/Controllers"
    _ "github.com/lib/pq"
)

// Configure logging - Warning and Error only
// Create a custom logger that only shows warnings and errors
func init() {
    // Set log flags to include timestamp
    log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
    // Note: Go's standard log package doesn't have severity levels,
    // but we can use log.Printf for warnings and log.Fatal/panic for errors
    // For production, consider using logrus or zap for proper log levels
}

func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}

func main() {
    databaseUrl := os.Getenv("DATABASE_URL")
    if databaseUrl == "" {
        log.Fatal("DATABASE_URL environment variable not set")
    }

    db, err := sql.Open("postgres", databaseUrl)
    if err != nil {
        log.Fatal("Failed to connect to database: ", err)
    }
    defer db.Close()

    if err := db.Ping(); err != nil {
        log.Fatal("Failed to ping database: ", err)
    }

    controller := controllers.NewTestController(db)
    mux := http.NewServeMux()

    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/" {
            http.NotFound(w, r)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        fmt.Fprintf(w, `{"message":"Backend API is running","status":"ok","swagger":"/swagger","api":"/api/test"}`)
    })

    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        fmt.Fprintf(w, `{"status":"healthy","service":"Backend API"}`)
    })

    // Swagger UI endpoint - serve interactive Swagger UI HTML page
    mux.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html")
        fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Backend API - Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin:0; background: #fafafa; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: "/swagger.json",
                dom_id: "#swagger-ui",
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>`)
    })

    // Swagger JSON endpoint - return OpenAPI spec as JSON
    mux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        fmt.Fprintf(w, `{
  "openapi": "3.0.0",
  "info": {
    "title": "Backend API",
    "version": "1.0.0",
    "description": "Go Backend API Documentation"
  },
  "paths": {
    "/api/test": {
      "get": {
        "summary": "Get all test projects",
        "responses": {
          "200": {
            "description": "List of test projects",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {
                    "$ref": "#/components/schemas/TestProjects"
                  }
                }
              }
            }
          }
        }
      },
      "post": {
        "summary": "Create a new test project",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/TestProjectsInput"
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Created test project",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/TestProjects"
                }
              }
            }
          }
        }
      }
    },
    "/api/test/{id}": {
      "get": {
        "summary": "Get test project by ID",
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "schema": {
              "type": "integer"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Test project found",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/TestProjects"
                }
              }
            }
          },
          "404": {
            "description": "Project not found"
          }
        }
      },
      "put": {
        "summary": "Update test project",
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "schema": {
              "type": "integer"
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/TestProjectsInput"
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Updated test project"
          },
          "404": {
            "description": "Project not found"
          }
        }
      },
      "delete": {
        "summary": "Delete test project",
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "schema": {
              "type": "integer"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Deleted successfully"
          },
          "404": {
            "description": "Project not found"
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "TestProjects": {
        "type": "object",
        "properties": {
          "Id": {
            "type": "integer"
          },
          "Name": {
            "type": "string"
          }
        }
      },
      "TestProjectsInput": {
        "type": "object",
        "required": ["Name"],
        "properties": {
          "Name": {
            "type": "string"
          }
        }
      }
    }
  }
}`)
    })

    // API routes handler function
    apiTestHandler := func(w http.ResponseWriter, r *http.Request) {
        path := r.URL.Path
        
        // Handle /api/test and /api/test/ (no ID) - normalize trailing slash
        if path == "/api/test" || path == "/api/test/" {
            switch r.Method {
            case "GET":
                controller.GetAll(w, r)
            case "POST":
                controller.Create(w, r)
            default:
                http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            }
            return
        }
        
        // Handle /api/test/:id
        if strings.HasPrefix(path, "/api/test/") {
            idStr := strings.TrimPrefix(path, "/api/test/")
            if idStr == "" {
                // Empty ID after /api/test/, treat as /api/test/
                switch r.Method {
                case "GET":
                    controller.GetAll(w, r)
                case "POST":
                    controller.Create(w, r)
                default:
                    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
                }
                return
            }
            
            id, err := strconv.Atoi(idStr)
            if err != nil {
                http.Error(w, "Invalid ID", http.StatusBadRequest)
                return
            }
            
            switch r.Method {
            case "GET":
                controller.GetById(w, r, id)
            case "PUT":
                controller.Update(w, r, id)
            case "DELETE":
                controller.Delete(w, r, id)
            default:
                http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            }
            return
        }
        
        http.NotFound(w, r)
    }

    // Register both /api/test and /api/test/ to handle trailing slashes
    mux.HandleFunc("/api/test", apiTestHandler)
    mux.HandleFunc("/api/test/", apiTestHandler)

    handler := corsMiddleware(mux)

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Printf("Server starting on 0.0.0.0:%s", port)
    log.Fatal(http.ListenAndServe("0.0.0.0:"+port, handler))
}
