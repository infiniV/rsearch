// rsearch Test Database - MongoDB
// Placeholder for future MongoDB translator development

db = db.getSiblingDB('rsearch_test');

db.products.insertMany([
  {
    product_code: '13w42',
    name: 'Widget Pro',
    description: 'High quality professional widget',
    rod_length: 150,
    price: 99.99,
    region: 'ca',
    in_stock: true,
    status: 'active',
    tags: ['premium', 'bestseller'],
    created_at: new Date()
  },
  {
    product_code: '13w43',
    name: 'Widget Basic',
    description: 'Standard everyday widget',
    rod_length: 200,
    price: 49.99,
    region: 'ny',
    in_stock: true,
    status: 'active',
    tags: ['basic'],
    created_at: new Date()
  },
  {
    product_code: '15x20',
    name: 'Gadget One',
    description: 'Multi-purpose gadget for home use',
    rod_length: 75,
    price: 79.99,
    region: 'cb',
    in_stock: true,
    status: 'active',
    tags: ['gadget'],
    created_at: new Date()
  }
]);

// Create indexes
db.products.createIndex({ product_code: 1 });
db.products.createIndex({ region: 1 });
db.products.createIndex({ price: 1 });
