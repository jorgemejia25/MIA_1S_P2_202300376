import { FileInfo } from "@/types/FileSystem";
import FileItem from "@/components/molecules/FileItem";
import React from "react";

interface FilesListProps {
  files: FileInfo[];
  onFileClick: (file: FileInfo) => void;
}

const FilesList: React.FC<FilesListProps> = ({ files, onFileClick }) => {
  return (
    <div className="flex flex-col divide-y divide-neutral-700/50">
      {files.map((file) => (
        <FileItem
          key={`${file.name}-${file.inodeId}`}
          file={file}
          onClick={onFileClick}
          viewMode="list"
        />
      ))}
    </div>
  );
};

export default FilesList;
