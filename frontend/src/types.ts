export interface Link {
  id: string;
  url: string;
  title: string | null;
  summary: string | null;
  content_text?: string | null;
  source_type: string | null;
  group_ids: string[];
  notes: string | null;
  status:
    | 'new'
    | 'fetching'
    | 'processing'
    | 'ready'
    | 'inaccessible'
    | 'paywalled'
    | 'failed';
  error_message: string | null;
  favicon_url: string | null;
  og_image_url: string | null;
  created_at: string;
  processed_at: string | null;
}

export interface Group {
  id: string;
  name: string;
  color: string;
  sort_order: number;
  created_at: string;
}

export interface GroupWithCount extends Group {
  link_count: number;
}

export interface Stats {
  total_links: number;
  by_status: Record<string, number>;
  by_group: Record<string, number>;
  by_source_type: Record<string, number>;
  ready_count: number;
  processing_count: number;
  failed_count: number;
  total_llm_cost: number;
}

export interface Config {
  fetcher_mode: 'simple' | 'crawl4ai';
}
