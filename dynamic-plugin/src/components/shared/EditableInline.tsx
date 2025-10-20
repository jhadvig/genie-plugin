import React, { useState, useEffect } from 'react';
import { Button, TextInput, Tooltip } from '@patternfly/react-core';
import { CheckIcon, TimesIcon } from '@patternfly/react-icons';

export type EditableInlineProps = {
  value: string;
  isTitle?: boolean;
  onConfirm: (updatedValue: string) => Promise<void> | void;
};

export function EditableInline({ value, isTitle, onConfirm }: EditableInlineProps) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(value);

  useEffect(() => {
    setDraft(value);
  }, [value]);

  return (
    <div style={{ display: 'flex', alignItems: 'center', margin: isTitle ? '0 0 10px 0' : '0 0 20px 0' }}>
      {!editing ? (
        <>
          <div
            role="button"
            tabIndex={0}
            onClick={() => setEditing(true)}
            onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { setEditing(true); e.preventDefault(); } }}
            style={{ cursor: 'pointer', display: 'inline-flex', alignItems: 'center' }}
          >
            {isTitle ? (
              <h2 style={{ margin: 0, fontSize: '24px', fontWeight: 'bold' }}>{value}</h2>
            ) : (
              <p style={{ margin: 0, color: '#666', fontSize: '14px' }}>{value}</p>
            )}
          </div>
        </>
      ) : (
        <div style={{ display: 'flex', alignItems: 'center', width: isTitle ? 480 : 640 }}>
          <TextInput
            value={draft}
            onChange={(_, newVal) => setDraft(newVal)}
            aria-label={isTitle ? 'Edit title' : 'Edit description'}
            placeholder={isTitle ? 'Enter title' : 'Enter description'}
          />
          <Tooltip content="Save">
            <Button
              variant="plain"
              aria-label="Save"
              onClick={async () => {
                await onConfirm(draft);
                setEditing(false);
              }}
              style={{ marginLeft: 8 }}
            >
              <CheckIcon />
            </Button>
          </Tooltip>
          <Tooltip content="Cancel">
            <Button
              variant="plain"
              aria-label="Cancel"
              onClick={() => {
                setDraft(value);
                setEditing(false);
              }}
            >
              <TimesIcon />
            </Button>
          </Tooltip>
        </div>
      )}
    </div>
  );
}


