
export type ImageStatus = {
  id: string;
  extension: string;
  process_type: number; // 0=Waiting 1=Processed 2=Deleted 3=Failed
};
