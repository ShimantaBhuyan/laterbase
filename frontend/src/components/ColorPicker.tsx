interface ColorPickerProps {
  value: string;
  onChange: (color: string) => void;
}

const presetColors = [
  '#5e6ad2',
  '#10b981',
  '#f59e0b',
  '#ef4444',
  '#8b5cf6',
  '#ec4899',
  '#06b6d4',
  '#84cc16',
  '#f97316',
  '#6b7280',
];

export default function ColorPicker({ value, onChange }: ColorPickerProps) {
  return (
    <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
      {presetColors.map((color) => {
        const isSelected = value === color;
        return (
          <button
            key={color}
            onClick={() => onChange(color)}
            title={color}
            style={{
              width: 28,
              height: 28,
              borderRadius: '50%',
              background: color,
              border: isSelected ? '2px solid var(--color-text-primary)' : '2px solid transparent',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              padding: 0,
              flexShrink: 0,
            }}
          >
            {isSelected && (
              <span style={{ color: '#fff', fontSize: 12, lineHeight: 1 }}>✓</span>
            )}
          </button>
        );
      })}
    </div>
  );
}
