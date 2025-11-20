# Assets Folder

This folder contains images uploaded for the gallery feature.

## Structure

- Gallery images are stored with naming convention: `gallery_<event_code>_<id>.ext`
- Example: `gallery_RAMADAN2025_img001.jpg`

## Supported Formats

- JPEG (.jpg, .jpeg)
- PNG (.png)
- GIF (.gif)

## Size Limits

- Maximum file size: 5MB per image

## Access

Images are served through the backend API endpoint: `/api/gallery/:id/image`
