SERVER TechStoreAPI {
    PORT: 5000
    DB: tech_store_db
}

ENTITY Category {
    name: string (unique)
    description: string (optional)
    active: bool = true
}

ENTITY Product {
    name: string
    sku: string (unique)
    price: float
    stock: int = 0
    category: Category
}

ENTITY Customer {
    username: string (unique)
    email: string (unique)
    vip: bool = false
}

ENTITY Order {
    order_code: string (unique)
    total: float = 0.0
    status: string = "pending"
    customer: Customer
    product: Product
}

ROUTE "/categories" {
    METHODS: GET, POST, DELETE
    TARGET: Category
}

ROUTE "/products" {
    METHODS: GET, POST, PUT, DELETE
    TARGET: Product
}

ROUTE "/customers" {
    METHODS: GET, POST, PUT
    TARGET: Customer
}

ROUTE "/orders" {
    METHODS: GET, POST, DELETE
    TARGET: Order
}