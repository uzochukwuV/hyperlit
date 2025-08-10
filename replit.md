# Hyperliquid Copy Trading Platform

## Overview

This is a comprehensive copy trading platform built for the Hyperliquid ecosystem that enables users to automatically follow and replicate trades from successful leaders. The platform consists of a Python-based analytics engine, FastAPI backend services, and a web frontend for user interaction. The system provides real-time trade monitoring, performance analysis, risk management, and machine learning-powered predictions to optimize trading strategies.

## User Preferences

Preferred communication style: Simple, everyday language.

## System Architecture

### Frontend Architecture
- **Technology**: Vanilla JavaScript with Bootstrap 5 for UI components
- **Structure**: Single-page application with section-based navigation
- **Real-time Updates**: WebSocket connections for live trade data and market updates
- **Charts**: Chart.js integration for performance visualization and analytics display
- **Responsive Design**: Mobile-first approach using Bootstrap's grid system

### Backend Architecture
- **Framework**: FastAPI for high-performance async API endpoints
- **Analytics Engine**: Comprehensive performance analysis system with caching mechanisms
- **ML Pipeline**: Scikit-learn based predictive models for leader performance forecasting
- **Database Layer**: AsyncPG for PostgreSQL connections with connection pooling
- **Real-time Processing**: Async/await patterns throughout for concurrent operations

### Data Models
- **Core Entities**: Leaders, Followers, Trades, Positions with comprehensive tracking
- **Performance Metrics**: Risk assessment, return calculations, drawdown analysis
- **Trade Management**: Order execution, status tracking, and portfolio optimization
- **Risk Management**: Multi-level risk assessment (LOW/MEDIUM/HIGH/EXTREME)

### API Design
- **RESTful Endpoints**: Structured around leader analytics, follower optimization, and trade management
- **Request/Response Models**: Pydantic schemas for data validation and serialization
- **Background Tasks**: Async processing for computationally intensive operations
- **CORS Configuration**: Configured for cross-origin requests during development

### Machine Learning Components
- **Predictive Models**: Random Forest and Gradient Boosting for performance prediction
- **Feature Engineering**: Automated extraction from historical trade data
- **Model Training**: Continuous learning from new trade data
- **Confidence Scoring**: Prediction reliability assessment

## External Dependencies

### Core Technologies
- **FastAPI**: Web framework for building APIs with automatic documentation
- **PostgreSQL**: Primary database for storing trade data, user information, and analytics
- **NumPy/Pandas**: Data processing and numerical computations for analytics
- **Scikit-learn**: Machine learning algorithms for predictive analytics
- **Chart.js**: Frontend charting library for data visualization

### Infrastructure Services
- **Hyperliquid API**: Direct integration for trade execution and market data
- **WebSocket Connections**: Real-time data streaming from Hyperliquid
- **Database Connection Pooling**: AsyncPG for efficient database operations

### Development Tools
- **Bootstrap 5**: Frontend UI framework for responsive design
- **Font Awesome**: Icon library for consistent UI elements
- **Uvicorn**: ASGI server for running the FastAPI application

### Blockchain Integration
- **Hyperliquid Network**: Native integration with Hyperliquid's Layer 1 blockchain
- **Asset Management**: Support for perpetuals, spot trading, and builder-deployed assets
- **Transaction Handling**: Nonce management and API wallet integration
- **Real-time Monitoring**: WebSocket feeds for live trade and position updates