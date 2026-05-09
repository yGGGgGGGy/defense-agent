import { useEffect, useRef, useCallback } from 'react';

interface SSEEvent {
  type: string;
  data: any;
}

export function useSSE(
  instanceId: string | null,
  onEvent: (event: SSEEvent) => void
) {
  const sourceRef = useRef<EventSource | null>(null);

  useEffect(() => {
    if (!instanceId) return;

    const url = `/api/v1/events?instance_id=${instanceId}`;
    const es = new EventSource(url);
    sourceRef.current = es;

    es.addEventListener('node_state', (e) => {
      try {
        const data = JSON.parse(e.data);
        onEvent({ type: 'node_state', data });
      } catch {}
    });

    es.addEventListener('instance_update', (e) => {
      try {
        const data = JSON.parse(e.data);
        onEvent({ type: 'instance_update', data });
      } catch {}
    });

    es.addEventListener('audit_append', (e) => {
      try {
        const data = JSON.parse(e.data);
        onEvent({ type: 'audit_append', data });
      } catch {}
    });

    es.onerror = () => {
      es.close();
    };

    return () => {
      es.close();
      sourceRef.current = null;
    };
  }, [instanceId, onEvent]);

  return sourceRef;
}
