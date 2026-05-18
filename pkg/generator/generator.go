package generator

import (
	"dsl-compiler/pkg/analyzer"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Generator struct {
	symbolTable *analyzer.SymbolTable
	outputDir   string
}

func New(st *analyzer.SymbolTable) *Generator {
	projectName := "generated_api"
	if st.Server != nil {
		projectName = st.Server.Name.Value
	}

	return &Generator{
		symbolTable: st,
		outputDir:   projectName,
	}
}

func (g *Generator) Generate() error {
	// 1. Crear estructura de carpetas
	folders := []string{
		g.outputDir,
		filepath.Join(g.outputDir, "models"),
		filepath.Join(g.outputDir, "routes"),
		filepath.Join(g.outputDir, "config"),
		filepath.Join(g.outputDir, "handlers"),
	}

	for _, folder := range folders {
		if err := os.MkdirAll(folder, 0755); err != nil {
			return fmt.Errorf("error al crear carpeta %s: %v", folder, err)
		}
	}

	if err := g.generateDBConfig(); err != nil {
		return err
	}
	if err := g.generateDockerCompose(); err != nil {
		return err
	}

	if err := g.generateModels(); err != nil {
		return err
	}
	if err := g.generateHandlers(); err != nil {
		return err
	}
	if err := g.generateRoutes(); err != nil {
		return err
	}
	if err := g.generateMain(); err != nil {
		return err
	}

	fmt.Printf("✅ Proyecto '%s' generado con éxito.\n", g.outputDir)
	fmt.Println("📦 Instalando dependencias automáticamente...")
	cmd := exec.Command("go", "mod", "init", g.outputDir)
	cmd.Dir = g.outputDir
	cmd.Run()

	cmdTidy := exec.Command("go", "mod", "tidy")
	cmdTidy.Dir = g.outputDir
	cmdTidy.Run()
	fmt.Println("\nPara comenzar:")
	fmt.Printf("  1. cd %s\n", g.outputDir)
	fmt.Println("  2. docker-compose up -d")
	fmt.Println("  3. go run main.go")

	return nil
}

func (g *Generator) generateDBConfig() error {
	path := filepath.Join(g.outputDir, "config", "database.go")
	port := 3307
	dbName := "api_db" // default
	if g.symbolTable.Server != nil {
		dbName = g.symbolTable.Server.Database
	}

	content := fmt.Sprintf(`package config

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	// 1. Conexión al servidor (sin base de datos) para asegurar que exista
	dsnRoot := "root:password@tcp(127.0.0.1:%d)/?charset=utf8mb4&parseTime=True&loc=Local"
	tempDB, err := gorm.Open(mysql.Open(fmt.Sprintf(dsnRoot)), &gorm.Config{})
	if err != nil {
		panic("Fallo al conectar al servidor MySQL: " + err.Error())
	}

	// 2. Crear la base de datos si no existe
	createDbQuery := "CREATE DATABASE IF NOT EXISTS %s"
	tempDB.Exec(createDbQuery)
	
	// Cerrar conexión temporal si fuera necesario, o simplemente proceder
    
	// 3. Conexión final a la base de datos específica
	dsn := "root:password@tcp(127.0.0.1:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local"
	database, err := gorm.Open(mysql.Open(fmt.Sprintf(dsn, %s)), &gorm.Config{})

	if err != nil {
		panic("Fallo al conectar a la base de datos: " + err.Error())
	}

	DB = database
}
`, port, dbName, port, "%s", "\""+dbName+"\"")

	return os.WriteFile(path, []byte(content), 0644)
}

func (g *Generator) generateDockerCompose() error {
	path := filepath.Join(g.outputDir, "docker-compose.yml")
	port := 3307
	dbName := "api_db"
	if g.symbolTable.Server != nil {
		dbName = g.symbolTable.Server.Database
	}

	content := fmt.Sprintf(`version: '3.8'
services:
  db:
    image: mysql:8.0
    container_name: %s_db
    restart: always
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: %s
    ports:
      - "%d:3306"
    volumes:
      - db_data:/var/lib/mysql

volumes:
  db_data:
`, g.outputDir, dbName, port)

	return os.WriteFile(path, []byte(content), 0644)
}

func (g *Generator) generateModels() error {
	caser := cases.Title(language.English)

	for name, entity := range g.symbolTable.Entities {
		path := filepath.Join(g.outputDir, "models", strings.ToLower(name)+".go")

		content := "package models\n\n"
		content += "import (\n\t\"github.com/google/uuid\"\n\t\"gorm.io/gorm\"\n)\n\n"
		content += "type " + name + " struct {\n"

		content += "  ID uuid.UUID `gorm:\"type:char(36);primaryKey\" json:\"id\"` \n"

		for _, field := range entity.Fields {
			goType := g.mapTypeToGo(field.Type)
			ptr := ""
			var gormModifiers []string
			binding := ""

			if field.IsOptional {
				ptr = "*"
			} else {
				gormModifiers = append(gormModifiers, "not null")
				binding = "binding:\"required\""
			}

			if field.IsUnique {
				gormModifiers = append(gormModifiers, "unique")
			}

			// --- INICIO DE LA MAGIA PARA VALORES POR DEFECTO ---
			if field.DefaultValue != nil {
				dv := field.DefaultValue.String()
				// GORM requiere que los strings en 'default' lleven comillas simples
				// Tu AST ya tiene tipos específicos, podemos usarlos:
				if field.Type == "string" {
					// Limpiamos las comillas del AST para el tag de GORM
					cleanVal := strings.Trim(dv, "\"")
					gormModifiers = append(gormModifiers, fmt.Sprintf("default:'%s'", cleanVal))
				} else {
					gormModifiers = append(gormModifiers, fmt.Sprintf("default:%s", dv))
				}
			}
			// --- FIN DE LA MAGIA ---

			if field.IsRelation {
				relID := caser.String(field.Name) + "ID"
				tagID := fmt.Sprintf("json:\"%s_id\"", strings.ToLower(field.Name))

				if len(gormModifiers) > 0 {
					tagID += fmt.Sprintf(" gorm:\"%s\"", strings.Join(gormModifiers, ";"))
				}
				if binding != "" {
					tagID += " " + binding
				}

				content += fmt.Sprintf("  %s %suuid.UUID `%s` \n", relID, ptr, tagID)
				content += fmt.Sprintf("  %s *%s `gorm:\"foreignKey:%s\" json:\"%s,omitempty\"` \n",
					caser.String(field.Name), field.Type, relID, strings.ToLower(field.Name))
			} else {
				tag := fmt.Sprintf("json:\"%s\"", strings.ToLower(field.Name))

				if len(gormModifiers) > 0 {
					tag += fmt.Sprintf(" gorm:\"%s\"", strings.Join(gormModifiers, ";"))
				}
				if binding != "" {
					tag += " " + binding
				}

				content += fmt.Sprintf("  %s %s%s `%s` \n", caser.String(field.Name), ptr, goType, tag)
			}
		}

		content += "}\n\n"

		content += fmt.Sprintf(`func (m *%s) BeforeCreate(tx *gorm.DB) (err error) {
  m.ID = uuid.New()
  return
}
`, name)

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) generateHandlers() error {
	caser := cases.Title(language.English)

	for name := range g.symbolTable.Entities {
		exportName := caser.String(name)
		fileName := strings.ToLower(name)
		path := filepath.Join(g.outputDir, "handlers", fileName+"_handler.go")

		allowedMethods := make(map[string]bool)
		for _, route := range g.symbolTable.Routes {
			if route.Target == name {
				for _, m := range route.Methods {
					allowedMethods[m] = true
				}
			}
		}

		// Inicio del archivo
		content := fmt.Sprintf(`package handlers

import (
	"net/http"
  "%s/config"
  "%s/models"
  "github.com/gin-gonic/gin"
  "gorm.io/gorm/clause"
)
`, g.outputDir, g.outputDir)

		// --- GET: Generamos GetAll y GetByID ---
		if allowedMethods["GET"] {
			content += fmt.Sprintf(`
// GetAll%s obtiene todos los registros
func GetAll%s(c *gin.Context) {
  var items []models.%s
  // Añadimos Preload para que la lista traiga relaciones
  config.DB.Preload(clause.Associations).Find(&items)
  c.JSON(http.StatusOK, items)
}

// Get%sByID obtiene un solo registro por ID
func Get%sByID(c *gin.Context) {
    id := c.Param("id")
    var entity models.%s
    
    if err := config.DB.Preload(clause.Associations).First(&entity, "id = ?", id).Error; err != nil {
        c.JSON(404, gin.H{"error": "Registro no encontrado"})
        return
    }
    c.JSON(200, entity)
}
`, exportName, exportName, exportName, exportName, exportName, exportName)
		}

		// --- POST: Generamos Create ---
		if allowedMethods["POST"] {
			content += fmt.Sprintf(`
// Create%s crea un nuevo registro
func Create%s(c *gin.Context) {
  var item models.%s
  
  // Usamos el decoder de Gin para forzar validación de campos
  if err := c.ShouldBindJSON(&item); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Campos inválidos o faltantes: " + err.Error()})
    return
  }
  
  if err := config.DB.Create(&item).Error; err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al guardar: " + err.Error()})
    return
  }
  c.JSON(http.StatusCreated, item)
}
`, exportName, exportName, exportName)
		}

		// --- PUT: Generamos Update ---
		if allowedMethods["PUT"] {
			content += fmt.Sprintf(`
// Update%s actualiza un registro
func Update%s(c *gin.Context) {
  var item models.%s
  id := c.Param("id")
  
  // 1. Verificar si el registro existe
  if err := config.DB.First(&item, "id = ?", id).Error; err != nil {
    c.JSON(http.StatusNotFound, gin.H{"error": "Registro no encontrado para actualizar"})
    return
  }

  // 2. Mapear los nuevos datos del JSON
  var input models.%s
  if err := c.ShouldBindJSON(&input); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: " + err.Error()})
    return
  }
  
  // 3. Intentar actualizar y CAPTURAR error de duplicados (Unique)
  if err := config.DB.Model(&item).Updates(input).Error; err != nil {
    // Si el error contiene "Duplicate entry", es una violación de unicidad
    c.JSON(http.StatusConflict, gin.H{
        "error": "Error de restricción: posiblemente el valor ya existe",
        "details": err.Error(),
    })
    return
  }
  
  c.JSON(http.StatusOK, item)
}
`, exportName, exportName, exportName, exportName)
		}

		// --- DELETE: Generamos Delete ---
		if allowedMethods["DELETE"] {
			content += fmt.Sprintf(`
// Delete%s borra un registro
func Delete%s(c *gin.Context) {
    id := c.Param("id")
    result := config.DB.Delete(&models.%s{}, "id = ?", id)
    
    if result.RowsAffected == 0 {
        c.JSON(404, gin.H{"error": "No se encontró el registro para eliminar"})
        return
    }
    c.JSON(200, gin.H{"message": "Eliminado exitosamente"})
}
`, exportName, exportName, exportName)
		}

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) generateRoutes() error {
	caser := cases.Title(language.English)
	path := filepath.Join(g.outputDir, "routes", "routes.go")

	content := fmt.Sprintf(`package routes

import (
	"%s/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

`, g.outputDir)

	for pathStr, route := range g.symbolTable.Routes {
		exportName := caser.String(route.Target)

		for _, method := range route.Methods {
			if method == "GET" || method == "POST" {
				prefix := "Create"
				if method == "GET" {
					prefix = "GetAll"
				}
				content += fmt.Sprintf("  r.%s(\"%s\", handlers.%s%s)\n",
					method, pathStr, prefix, exportName)
			}

			if route.HasIndividualActions && (method == "GET" || method == "PUT" || method == "DELETE") {
				handlerName := ""
				switch method {
				case "GET":
					handlerName = "Get" + exportName + "ByID"
				case "PUT":
					handlerName = "Update" + exportName
				case "DELETE":
					handlerName = "Delete" + exportName
				}

				if handlerName != "" {
					content += fmt.Sprintf("	r.%s(\"%s/:id\", handlers.%s)\n",
						method, pathStr, handlerName)
				}
			}
		}
	}

	content += "\n	return r\n}"

	return os.WriteFile(path, []byte(content), 0644)
}

func (g *Generator) generateMain() error {
	path := filepath.Join(g.outputDir, "main.go")

	port := "8080"
	if g.symbolTable.Server != nil {
		port = strconv.Itoa(g.symbolTable.Server.Port)
	}

	content := fmt.Sprintf(`package main

import (
	"%s/config"
	"%s/routes"
	"%s/models"
	"fmt"
)

func main() {
	// 1. Inicializar conexión a MySQL
	config.ConnectDatabase()

	// 2. Ejecutar Migraciones Automáticas
	// Esto crea las tablas y relaciones en la DB según tus modelos
	config.DB.AutoMigrate(
`, g.outputDir, g.outputDir, g.outputDir)

	for name := range g.symbolTable.Entities {
		content += fmt.Sprintf("		&models.%s{},\n", name)
	}

	content += fmt.Sprintf(`	)

	// 3. Arrancar el servidor Gin
	r := routes.SetupRouter()
	fmt.Println("🚀 Servidor DSL corriendo en el puerto: %s")
	r.Run(":%s")
}
`, port, port)

	return os.WriteFile(path, []byte(content), 0644)
}

// Función auxiliar para mapear tipos del DSL a Go
func (g *Generator) mapTypeToGo(dslType string) string {
	switch dslType {
	case "int":
		return "int"
	case "string":
		return "string"
	case "float":
		return "float64"
	case "bool":
		return "bool"
	default:
		return dslType // Si es una entidad, devolvemos el mismo nombre
	}
}
