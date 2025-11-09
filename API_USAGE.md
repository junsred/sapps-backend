# Humanization API Usage

This document explains how to use the AI text humanization and detection API endpoints.

## Endpoints

### POST /humanizations
Creates a new humanization task that will process the input text with ChatGPT and check AI detection percentages.

**Request:**
```json
{
  "input_text": "Artificial intelligence has revolutionized the way we approach data analysis. Machine learning algorithms can process vast amounts of information with unprecedented efficiency and accuracy."
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Humanization started. Check back for results.",
  "humanization": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "input_text": "Artificial intelligence has revolutionized...",
    "status": "processing",
    "user_id": "user123"
  }
}
```

### GET /humanizations
Retrieves all humanization tasks for the authenticated user, including completed results with AI detection scores.

**Response:**
```json
{
  "status": "success",
  "humanizations": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "input_text": "Original AI-generated text...",
      "output_text": "Humanized version of the text...",
      "original_detection": "[{\"service\":\"Sapling\",\"score\":0.85,\"confidence\":\"Likely AI\"},{\"service\":\"GPTZero\",\"score\":0.78,\"confidence\":\"Likely AI\"}]",
      "final_detection": "[{\"service\":\"Sapling\",\"score\":0.23,\"confidence\":\"Likely Human\"},{\"service\":\"GPTZero\",\"score\":0.31,\"confidence\":\"Likely Human\"}]",
      "improvement_score": 72.5,
      "status": "completed",
      "created_date": "2024-01-15T10:30:00Z",
      "user_id": "user123"
    }
  ]
}
```

## Features

### AI Detection Services
The system integrates with multiple AI detection APIs:
- **Sapling.ai**: Industry-leading AI content detection
- **GPTZero**: Specialized in detecting GPT-generated content  
- **Additional services**: Extensible architecture for more detectors

### Humanization Process
1. **Original Analysis**: Checks AI detection scores on input text
2. **ChatGPT Humanization**: Uses advanced prompts to make text more natural
3. **Final Analysis**: Re-checks AI detection scores on sappsd text
4. **Improvement Calculation**: Shows percentage improvement in human-like appearance

### Response Format
- **Scores**: Range from 0.0 (definitely human) to 1.0 (definitely AI)
- **Confidence Levels**: 
  - "Likely Human" (< 0.3)
  - "Uncertain" (0.3-0.7) 
  - "Likely AI" (> 0.7)
- **Improvement Score**: Percentage reduction in AI detection confidence

## Authentication
All endpoints require user authentication via JWT token in the Authorization header.

## Status Tracking
- **processing**: Humanization is in progress
- **completed**: Successfully processed with results
- **failed**: Processing encountered an error

## Example Usage

### 1. Submit text for humanization
```bash
curl -X POST https://api.example.com/humanizations \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"input_text": "Your AI-generated text here..."}'
```

### 2. Check results
```bash
curl -X GET https://api.example.com/humanizations \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

The system processes humanization asynchronously, so you may need to poll the GET endpoint until the status changes from "processing" to "completed". 