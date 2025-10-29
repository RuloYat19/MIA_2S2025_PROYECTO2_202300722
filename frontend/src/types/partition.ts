export interface FileOrFolder {
  name: string;
  type: "file" | "folder";
  children?: FileOrFolder[];
}

export interface TreeNode {
  name: string;
  isDir: boolean;
  children?: TreeNode[];
}

export interface Partition {
  id: string;
  isMounted: boolean;
  name: string;
  size: number;
  start: number;
}