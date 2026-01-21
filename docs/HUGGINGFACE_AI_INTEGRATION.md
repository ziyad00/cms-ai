# Hugging Face AI Integration

This document describes the Hugging Face AI orchestrator implementation for generating TemplateSpec objects from natural language prompts.

## Overview

The AI integration replaces the previous stub template generation with actual AI-powered template creation using Hugging Face's free-tier models. It provides intelligent template generation based on user prompts with support for brand kits, languages, tones, and RTL layouts.

## Architecture

### Components

1. **HuggingFaceClient** - HTTP client for Hugging Face Inference API
2. **Orchestrator** - Manages AI generation and repair workflows  
3. **AIService** - High-level service with store integration and metering
4. **API Integration** - Updated `POST /v1/templates/generate` endpoint

### Data Flow

```
User Prompt → API Endpoint → AIService → Orchestrator → HuggingFaceClient → AI Model → TemplateSpec
```

## Configuration

### Environment Variables

- `HUGGINGFACE_API_KEY` - Hugging Face API key (required for AI generation)
- `HUGGINGFACE_MODEL` - Model to use (default: `mistralai/Mixtral-8x7B-Instruct-v0.1`)

### Setup

1. Create a Hugging Face account at https://huggingface.co
2. Generate an API token from https://huggingface.co/settings/tokens
3. Set the environment variable: `export HUGGINGFACE_API_KEY=your_token_here`

## Usage

### API Endpoint

```bash
POST /v1/templates/generate
Content-Type: application/json
X-User-ID: user-id
X-Org-ID: org-id
X-User-Role: editor

{
  "prompt": "Create a modern tech startup pitch deck template",
  "name": "Startup Pitch Deck",
  "brandKitId": "bk-123", // optional
  "language": "English",   // optional
  "tone": "Professional",  // optional
  "rtl": false            // optional
}
```

### Response

```json
{
  "template": { /* template metadata */ },
  "version": { /* version with generated spec */ },
  "aiResponse": {
    "model": "mistralai/Mixtral-8x7B-Instruct-v0.1",
    "tokenUsage": 245,
    "cost": 0.00123,
    "timestamp": "2026-01-15T10:30:00Z"
  }
}
```

## Prompt Engineering

### System Prompt Structure

- **Instructions**: Clear description of TemplateSpec requirements
- **Schema**: Exact JSON structure with field descriptions
- **Rules**: Geometry constraints, naming conventions, design principles
- **Context**: Language, tone, RTL, brand kit integration
- **Examples**: Few-shot demonstrations for better output quality

### Few-Shot Examples

The system includes curated examples for:
- Modern tech startup templates
- Corporate business templates
- Creative presentation templates

### Brand Kit Integration

When a brand kit ID is provided, the AI:
1. Loads the brand kit tokens from the store
2. Extracts colors, fonts, and brand elements
3. Incorporates them into the generated template
4. Maintains brand consistency while meeting the prompt requirements

## Model Configuration

### Default Model: Mixtral-8x7B-Instruct-v0.1

- **Free tier**: Yes (rate limited)
- **Token limit**: 32k context
- **Quality**: Excellent for structured JSON generation
- **Parameters**:
  - Temperature: 0.7 (balanced creativity)
  - Top-p: 0.9 (focused sampling)
  - Max tokens: 2048 (sufficient for templates)

### Alternative Models

The system supports any Hugging Face text generation model. Configure via:

```bash
export HUGGINGFACE_MODEL="meta-llama/Llama-2-7b-chat-hf"
```

## Error Handling & Fallbacks

### Primary Fallback

If AI generation fails, the system gracefully falls back to the original stub template spec with:
- Professional color scheme
- Title and subtitle placeholders
- Safe geometry constraints

### Repair Loop

The orchestrator includes a repair mechanism for invalid specs:
1. Detects validation errors
2. Generates repair prompt with error details
3. Re-runs generation with corrections
4. Validates repaired output

## Cost Management

### Token Usage Tracking

- Input tokens: System prompt + user prompt
- Output tokens: Generated TemplateSpec JSON
- Total usage tracked per organization

### Pricing

Based on Mixtral pricing (as of 2024):
- Input: ~$0.50 per 1M tokens
- Output: ~$1.50 per 1M tokens
- Typical generation: $0.001-0.005

### Quotas

AI generation counts against the same quota as template generation:
- Default: 50 per month per organization
- Configurable via `GENERATE_LIMIT_PER_MONTH`

## Testing

### Unit Tests

```bash
go test ./internal/ai/... -v
```

### Integration Tests

```bash
go test ./internal/api/... -run TestGenerateTemplate -v
```

### Mock Testing

The system includes comprehensive mock services for testing without API calls:
- Mock HuggingFace client
- Mock AI service
- Mock store implementations

## Security

### API Key Management

- Stored in environment variables only
- Never logged or exposed in responses
- Validated on client initialization

### Input Validation

- Prompt length limits (via HTTP MaxBytesReader)
- Malicious input sanitization
- Structured output validation

### Output Validation

- JSON schema validation against TemplateSpec
- Geometry constraint verification
- Safe margin enforcement

## Monitoring

### Metrics Tracked

- Token usage per organization/model
- Generation success/failure rates
- Response times
- Cost tracking

### Logs Generated

- Generation requests with prompts
- Model responses and token counts
- Validation errors and repairs
- Cost and quota impacts

## Performance

### Response Times

- Typical: 2-10 seconds depending on prompt complexity
- Timeout: 60 seconds (configurable)
- Rate limiting: Handled by Hugging Face

### Caching

Consider implementing:
- Response caching for similar prompts
- Brand kit caching
- Template pattern optimization

## Future Enhancements

### Planned Features

1. **Multiple model support** - Choose from different AI models
2. **Fine-tuning** - Custom models trained on presentation templates
3. **Async generation** - Background processing for complex prompts
4. **Template variations** - Generate multiple options
5. **Style transfer** - Apply different visual styles to templates

### Advanced Prompting

1. **Industry-specific prompts** - Healthcare, finance, education templates
2. **Multi-language support** - Generate in any language
3. **Cultural adaptation** - Region-specific design patterns
4. **Accessibility compliance** - WCAG-compliant template generation

## Troubleshooting

### Common Issues

1. **API Key Not Found**
   - Check `HUGGINGFACE_API_KEY` environment variable
   - Verify key is valid and not expired

2. **Rate Limiting**
   - Free tier has request limits
   - Consider upgrading plan for production use

3. **Invalid JSON Output**
   - Model sometimes generates malformed JSON
   - Repair loop handles most cases
   - Falls back to stub template if unrecoverable

4. **High Costs**
   - Monitor token usage
   - Adjust prompt complexity
   - Consider input token optimization

### Debug Mode

Set `LOG_LEVEL=debug` to see:
- Full prompts sent to model
- Raw responses from API
- Validation error details
- Repair attempt logs