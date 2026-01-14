export function stubTemplateSpec() {
  return {
    tokens: {
      colors: {
        primary: '#3366FF',
        background: '#FFFFFF',
        text: '#111111',
      },
    },
    constraints: {
      safeMargin: 0.05,
    },
    layouts: [
      {
        name: 'Title / Hero',
        placeholders: [
          {
            id: 'title',
            type: 'text',
            geometry: { x: 0.1, y: 0.2, w: 0.8, h: 0.2 },
          },
          {
            id: 'subtitle',
            type: 'text',
            geometry: { x: 0.1, y: 0.45, w: 0.8, h: 0.15 },
          },
        ],
      },
    ],
  }
}
