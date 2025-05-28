# Work Log Management System

A comprehensive work log management application built with modern web technologies, featuring advanced search capabilities and visual analytics for tracking tasks, bugs, and stories.

![worklog](https://github.com/user-attachments/assets/ce433a4a-119f-466b-86cb-2ef22b18fb11)

![worklog](https://github.com/user-attachments/assets/b5c0e697-1280-4348-b8c7-f3c853d883d1)

![output-0](https://github.com/user-attachments/assets/3af329ba-9d32-4486-813b-9e509c4cb7e1)

![output-1](https://github.com/user-attachments/assets/03a95c5a-8f46-4438-8dff-9816dae22ab9)

![output-2](https://github.com/user-attachments/assets/4485d822-23d2-4636-95a6-81d1a3dd1ce5)

![output-3](https://github.com/user-attachments/assets/93297ed4-4db8-4790-b60b-59bb6de6d07f)

![output-4](https://github.com/user-attachments/assets/d046660a-9bc6-46f8-b7bc-d841e220eb39)


## 🚀 Features

### Core Functionality
- **Task Management**: Create, update, and track tasks with different types (Task, Bug, Story)
- **Status Tracking**: Monitor progress with status categories (Pending, Progress, Staging, PR, Backlog)
- **Priority System**: Assign priority levels to organize work effectively
- **Rich Notes**: Add detailed notes and descriptions for each work item

### Advanced Search Capabilities
- **Fuzzy Search**: Levenshtein distance-based search for finding similar task names
- **Trigram Similarity**: PostgreSQL trigram matching for flexible text search
- **Full-Text Search**: Lexeme-based search using PostgreSQL's text search vectors
- **Phonetic Search**: Metaphone algorithm for sound-based matching
- **Real-time Search**: Instant results as you type

### Analytics & Visualization
- **Task Completion Timeline**: Line chart showing completion trends over time
- **Status Distribution**: Pie chart visualization of task status breakdown
- **Task Type Analysis**: Visual breakdown of work item types
- **Interactive Dashboard**: Real-time metrics and insights

## 🛠️ Tech Stack

- **Backend**: Go (Golang) with net/http
- **Database**: PostgreSQL with advanced extensions
- **Frontend**: Vue.js 3
- **Search Extensions**: 
  - `fuzzystrmatch` (Levenshtein distance)
  - `pg_trgm` (Trigram matching)
  - Built-in full-text search capabilities

## 🏗️ Database Architecture

### Core Table Structure
```sql
-- Main logs table with comprehensive task tracking
CREATE TABLE logs (
    id SERIAL PRIMARY KEY,
    task_name VARCHAR(255) NOT NULL,
    task_type VARCHAR(50),
    task_status VARCHAR(50),
    priority INTEGER,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    -- Auto-generated search vector for full-text search
    task_name_vector TSVECTOR GENERATED ALWAYS AS (TO_TSVECTOR('english', task_name)) STORED
);
```

### Search Optimization
- **GIN Index**: Optimized indexing for full-text search vectors
- **Generated Columns**: Automatic tsvector generation for efficient text search
- **Multiple Search Algorithms**: Levenshtein, Trigram, and Lexeme-based searching

## 🔍 Search Features Deep Dive

### 1. Levenshtein Distance (Edit Distance)
```sql
-- Find tasks with similar names (fuzzy matching)
SELECT task_name, LEVENSHTEIN('write docs', LOWER(task_name)) AS distance
FROM logs
ORDER BY distance ASC;
```

### 2. Trigram Similarity
```sql
-- Similarity-based search with percentage matching
SELECT notes, SIMILARITY('module implementation', notes) AS similarity
FROM logs
WHERE notes % 'module implementation'
ORDER BY similarity DESC;
```

### 3. Full-Text Search
```sql
-- Lexeme-based search with ranking
SELECT task_name, ts_rank(task_name_vector, to_tsquery('system | status')) as rank
FROM logs
WHERE task_name_vector @@ to_tsquery('system | status')
ORDER BY rank DESC;
```

### 4. Phonetic Search (Metaphone)
```sql
-- Sound-based matching for misspelled queries
SELECT task_name FROM logs
WHERE METAPHONE(task_name::VARCHAR, 4) = METAPHONE('progres', 4);
```

## 📊 Analytics Features

- **Task Completion Trends**: Track productivity over time
- **Status Distribution**: Visual breakdown of work pipeline
- **Type Analysis**: Understand work composition (Tasks vs Bugs vs Stories)
- **Priority Insights**: Monitor high-priority item completion

## 🚀 Getting Started

### Prerequisites
- Go 1.19+
- PostgreSQL 12+
- Node.js 16+
- Vue CLI

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd worklog-project
   ```

2. **Set up PostgreSQL**
   ```sql
   -- Create database
   CREATE DATABASE worklog;
   
   -- Install required extensions
   CREATE EXTENSION fuzzystrmatch;
   CREATE EXTENSION pg_trgm;
   ```

3. **Backend Setup**
   ```bash
   # Install Go dependencies
   go mod download
   
   # Set up environment variables
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_NAME=worklog
   export DB_USER=your_username
   export DB_PASSWORD=your_password
   
   # Run the server
   go run main.go
   ```

4. **Frontend Setup**
   ```bash
   # Install dependencies
   npm install
   
   # Start development server
   npm run serve
   ```

## 🎯 Usage

### Adding Tasks
1. Use the search interface to add new tasks
2. Specify task type (Task, Bug, Story)
3. Set priority and status
4. Add detailed notes

### Searching
- **Quick Search**: Type in the search box for instant results
- **Advanced Filters**: Use the interface to filter by status, type, and priority
- **Fuzzy Matching**: The system automatically handles typos and similar terms

### Analytics
- View the dashboard for real-time insights
- Track completion trends over time
- Analyze work distribution and bottlenecks

## 🧠 Learning Outcomes

This project demonstrates proficiency in:

- **Database Design**: Advanced PostgreSQL features and indexing strategies
- **Search Algorithms**: Multiple text search implementations
- **Go Programming**: HTTP servers, database integration, and pointer management
- **Frontend Development**: Vue.js 3 reactive interfaces
- **Data Visualization**: Chart.js integration for analytics
- **Performance Optimization**: Efficient database queries and indexing

## 📈 Advanced Features

- **B-Tree Indexing**: Self-balancing tree structures for optimal search performance
- **Generated Columns**: Automatic tsvector maintenance
- **Common Table Expressions**: Complex query optimization
- **Window Functions**: Advanced analytics with OVER() clauses

## 🤝 Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues for bugs and feature requests.

## 📝 License

This project is open source and available under the [MIT License](LICENSE).

---

**Built with ❤️ using Go, PostgreSQL, and Vue.js**
