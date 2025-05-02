import { FileInfo } from "@/types/FileSystem";
import FileItem from "@/components/molecules/FileItem";
import React from "react";

interface FilesGridProps {
  files: FileInfo[];
  onFileClick: (file: FileInfo) => void;
}

const FilesGrid: React.FC<FilesGridProps> = ({ files, onFileClick }) => {
  return (
    <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
      {files.map((file) => (
        <FileItem
          key={`${file.name}-${file.inodeId}`}
          file={file}
          onClick={onFileClick}
          viewMode="grid"
        />
      ))}
    </div>
  );
};

export default FilesGrid;
