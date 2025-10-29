export interface Disk{
    fileName : string;
    filePath : string;
}

export interface DiskState{
    disks: Disk[];
    loading: boolean;
    error: string | null;
    addDisk: (filePath: string) => void;
    addDisksFromFolder: (filePaths: string[]) => void;
    setLoading: (isLoading: boolean) => void;
    clearDisks: () => void;
    removeDisk: (diskIndex: number) => void;
}