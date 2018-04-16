class InitializeIndexerTables < ActiveRecord::Migration
  def change
    create_table :block_headers do |t|
      t.string :hash, :limit => 64, :null => false
      t.string :parent_hash, :limit => 64, :null => false
      t.string :uncle_hash, :limit => 64, :null => false
      t.string :coinbase, :limit => 40, :null => false
      t.string :root, :limit => 64, :null => false
      t.string :tx_hash, :limit => 64, :null => false
      t.string :receipt_hash, :limit => 64, :null => false
      t.binary :bloom
      t.integer :difficulty, :limit => 8, :null => false
      t.integer :number, :limit => 8, :null => false
      t.column :gas_limit, 'BIGINT UNSIGNED'
      t.column :gas_used, 'BIGINT UNSIGNED'
      t.column :time, 'BIGINT UNSIGNED'
      t.binary :extra_data
      t.string :mix_digest
      t.binary :nonce
    end
    add_index :block_headers, :number, :unique => true

    create_table :transactions do |t|
      t.string :hash, :limit => 64
      t.string :block_hash, :limit => 64
      t.string :from, :limit => 40
      t.string :to, :limit => 40
      t.binary :nonce
      t.integer :gas_price, :limit => 8
      t.column :gas_limit, 'BIGINT UNSIGNED'
      t.integer :amount, :limit => 8
      t.binary :payload
      t.integer :v, :limit => 8
      t.integer :s, :limit => 8
      t.integer :r, :limit => 8
    end
    add_index :transactions, :hash, :unique => true

    create_table :transaction_receipts do |t|
      t.binary :root, :limit => 64
      t.column :status, 'INT UNSIGNED'
      t.column :cumulative_gas_used, 'BIGINT UNSIGNED'
      t.binary :bloom
      t.string :tx_hash, :limit => 64
      t.string :contract_address, :limit => 40
      t.column :gas_used, 'BIGINT UNSIGNED'
    end
    add_index :transaction_receipts, :tx_hash, :unique => true

    # Use state_blocks to store the blocks which we updated them to our accounts and contracts table
    create_table :state_blocks do |t|
      t.integer :number, :limit => 8, :null => false
    end
    add_index :state_blocks, :number, :unique => true

    create_table :contract_code do |t|
      t.binary :address, :limit => 20, :null => false
      t.binary :hash, :limit => 32, :null => false
      t.text :code, :limit => 10.megabyte, :null => false
    end
    add_index :contract_code, :address, :unique => true

    create_table :accounts do |t|
      t.binary :address, :limit => 20, :null => false
      t.integer :block_number, :limit => 8, :null => false
      t.string :balance, :limit => 32, :null => false
      t.integer :nonce, :limit => 8
    end
    add_index :accounts, :address
    add_index :accounts, [:address, :block_number], :unique => true

    create_table :contracts do |t|
      t.binary :address, :limit => 20, :null => false
      t.integer :block_number, :limit => 8, :null => false
      t.string :balance, :limit => 32, :null => false
      t.integer :nonce, :limit => 8, :null => false
      t.binary :root, :limit => 32, :null => false
      t.binary :storage, :limit => 10.megabyte, :null => false
    end
    add_index :contracts, :address
    add_index :contracts, [:address, :block_number], :unique => true

    # TODO: Add foreign keys?
  end
end
