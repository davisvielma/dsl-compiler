# Orion DSL 🚀

**Orion DSL** es un Lenguaje de Dominio Específico (DSL) diseñado para simplificar y automatizar la creación de microservicios y APIs REST profesionales en Go. En lugar de escribir manualmente cientos de líneas de código repetitivo (_boilerplate_) para operaciones CRUD, conexión a bases de datos y configuraciones de contenedores, Orion permite definir todo el sistema de manera declarativa y limpia en un único archivo de especificación.

Este proyecto ha sido desarrollado como proyecto final para la materia **Compiladores** en la **Universidad de Los Andes (ULA)**, Mérida, Venezuela.

## 🛠 Características del Lenguaje

- **Enfoque Declarativo de Infraestructura:** Define servidores, entidades y rutas de red sin implementar lógica de bajo nivel.
- **Análisis Semántico Estricto:** Validación de la existencia de entidades referenciadas, compatibilidad de tipos y consistencia de contratos de API antes de generar código.
- **Seguridad y Persistencia Nativa:**
  - Generación automática de llaves primarias basadas en **UUID** (`char(36)`).
  - Soporte nativo para relaciones de base de datos e integridad referencial.
- **Lógica de Atributos Predictiva:**
  - **Obligatoriedad por defecto:** Todo campo es requerido (`NOT NULL` en base de datos y validación `binding:"required"` en Go) a menos que se indique lo contrario.
  - **`optional`**: Implementado mediante el uso de **punteros** (`*type`) en Go para diferenciar un valor no enviado de uno vacío o cero.
  - **`unique`**: Restricción de unicidad estricta integrada directamente en el motor de base de datos.
- **Generación de Entorno de Ejecución:** Emisión automática de código Go limpio utilizando **Gin Gonic** para enrutamiento, **GORM** para persistencia relacional, y un archivo `docker-compose.yml` con MySQL 8.0 configurado.

## 📂 Estructura del Repositorio

La arquitectura del proyecto separa estrictamente la fase de análisis del compilador de la fase de generación de código:

```text
/dsl-compiler
├── main.go              # Punto de entrada del compilador Orion
├── api.dsl              # Archivo de especificación DSL (Entorno de pruebas)
├── pkg/
│   └── analyzer/        # Análisis semántico y Tabla de Símbolos
│   └── generator/       # Fase de Síntesis (Back-end del compilador)
│   ├── lexer/           # Analizador léxico (Extracción de tokens)
│   ├── parser/          # Analizador sintáctico (Construcción del AST)
│   └── ast/             # ast (Árbol de Sintaxis Abstracta)
```

## 🚀 Guía de Inicio Rápido

- **Requisitos Previos**
  - Go versión 1.22 o superior.
  - Docker y Docker Compose instalados y activos.

### Instrucciones para generar y levantar la API

```bash
# Clonar el proyecto
git clone https://github.com/davisvielma/dsl-compiler.git
cd dsl-compiler

# Inicializar el módulo (si no existe el archivo go.mod)
go mod init orion-dsl

# Descargar dependencias del compilador
go mod tidy

# Compilar el proyecto
go run main.go

# Entrar a la carpeta generada (ajusta 'output' al nombre real que crea el compilador)
cd output

# Levantar la base de datos MySQL con Docker
docker-compose up -d

# Ejecutar el servidor
go run main.go
```

## 🏗 Pipeline del Compilador

El ciclo de vida del compilador de Orion DSL se ejecuta de manera secuencial a través de las siguientes fases lógicas:

1. **Fase de Análisis Léxico y Sintáctico:** Escanea el archivo fuente api.dsl agrupando caracteres en componentes sintácticos (tokens) y organizándolos en una estructura de árbol conocida como Árbol de Sintaxis Abstracta (AST).

2. **Fase de Análisis Semántico:** Pobla la Tabla de Símbolos y valida las declaraciones. Asegura que los tipos mapeados sean válidos, que los campos de relación apunten a entidades existentes y que las rutas HTTP utilicen métodos válidos vinculados a entidades reales.

3. **Fase de Síntesis (Generación):** Traduce el AST validado en código fuente estructurado de Go. Crea las subcarpetas del servidor (models, handlers, routes, config), configura el pool de conexiones de base de datos relacional y escribe la automatización de infraestructura de Docker.
