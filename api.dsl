/* PROYECTO: Generador de API REST - DSL Profesional
   PRUEBA: Verificación de Lexer (Líneas, Números y Case Insensitivity)
   /* Este es un comentario anidado real */
*/

// Probando normalización: Server, SERVER y server deben reconocerse igual
SERVER MiTienda_2026 {
    PORT: 9090
    DB: sqlite3
}

/* Entidad: Order
   Probando Identificadores con números y guiones bajos
*/
ENTITY Order_v1 {
    id: int
    customer_id: int
    total_price: float
    status: string // "pending", "shipped", "delivered"
    //@
}

// Probando comentarios de una sola línea pegados a código
entity Category { // Categorización de productos
    id: int
    name: string
}

/* Probando rutas con números, múltiples slashes
   y la normalización de 'Route', 'Methods' y 'Target'
*/
Route /api/v2/orders/report {
    METHODS: GET
    TARGET: Order_v1
}

route /api/v1/auth/register {
    methods: POST, PUT
    target: User
}

/* CASOS DE PRUEBA DE ERROR (Descomenta uno a la vez para probar el errorf)
*/

// Error de carácter inesperado:
 # 

// Error de comentario multilínea sin cerrar:
// /* Este comentario se queda abierto hasta el fin del mundo...
@