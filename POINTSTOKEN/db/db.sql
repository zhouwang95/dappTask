CREATE TABLE IF NOT EXISTS chains (
      id INT AUTO_INCREMENT PRIMARY KEY,
      name VARCHAR(50) NOT NULL UNIQUE,
      chain_id BIGINT NOT NULL UNIQUE,
      contract_addr VARCHAR(66) NOT NULL,
      last_processed_block BIGINT NOT NULL DEFAULT 0,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS user_balances (
     id INT AUTO_INCREMENT PRIMARY KEY,
     chain_id BIGINT NOT NULL,
     user_addr VARCHAR(66) NOT NULL,
     balance DECIMAL(50, 0) NOT NULL DEFAULT 0,
     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
     UNIQUE KEY unique_user_chain (chain_id, user_addr)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS balance_changes (
       id INT AUTO_INCREMENT PRIMARY KEY,
       chain_id BIGINT NOT NULL,
       user_addr VARCHAR(66) NOT NULL,
       transaction_hash VARCHAR(66) NOT NULL,
       block_number BIGINT NOT NULL,
       change_amount DECIMAL(50, 0) NOT NULL,
       balance_after DECIMAL(50, 0) NOT NULL,
       event_type VARCHAR(20) NOT NULL,
       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE IF NOT EXISTS user_points (
       id INT AUTO_INCREMENT PRIMARY KEY,
       chain_id BIGINT NOT NULL,
       user_addr VARCHAR(66) NOT NULL UNIQUE,
       total_points DECIMAL(50, 0) NOT NULL DEFAULT 0,
       updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
       UNIQUE KEY unique_user_points (chain_id, user_addr)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE IF NOT EXISTS points_calculations (
       id INT AUTO_INCREMENT PRIMARY KEY,
       chain_id BIGINT NOT NULL,
       user_addr VARCHAR(66) NOT NULL,
       calculated_at TIMESTAMP NOT NULL,
       balance DECIMAL(50, 0) NOT NULL,
       points_added DECIMAL(50, 0) NOT NULL,
       total_points_after DECIMAL(50, 0) NOT NULL,
       KEY idx_user_addr (user_addr),
       KEY idx_calculated_at (calculated_at),
       UNIQUE KEY unique_points_calculations (chain_id, user_addr, calculated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;