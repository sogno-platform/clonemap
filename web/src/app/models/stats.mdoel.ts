
export interface StatsInfo  { 
  max: number;
  min: number;
  count: number;
  average: number;
  list: BehStats[]
}

interface BehStats {
  start: string;
  end: string;
  duration: number;
}