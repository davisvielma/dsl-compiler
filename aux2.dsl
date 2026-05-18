SERVER BookStoreAPI {
    PORT: 3000
    DB: bookstore_db
}

ENTITY Author {
    name: string (unique, optional)
    nationality: string = "Venezolano" 
    age: int
}

ENTITY Book {
    title: string
    pages: int (unique)
    price: float (optional)
    author: Author
}

ROUTE "/authors" {
    METHODS: GET, POST, DELETE
    TARGET: Author
}

ROUTE "/books" {
    METHODS: GET, POST, PUT, DELETE
    TARGET: Book
}