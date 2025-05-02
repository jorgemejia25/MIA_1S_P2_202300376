import { HiOutlineDocumentText, HiOutlineFolder } from "react-icons/hi";

import React from "react";

interface OperationIconProps {
  operation: string;
  className?: string;
}

const OperationIcon: React.FC<OperationIconProps> = ({
  operation,
  className = "w-5 h-5",
}) => {
  switch (operation) {
    case "mkdir":
      return <HiOutlineFolder className={className} />;

    case "mkfile":
    case "append":
    case "rename":
      return <HiOutlineDocumentText className={className} />;

    case "chmod":
      return (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className={className}
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
          />
        </svg>
      );

    case "rm":
      return (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className={className}
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
          />
        </svg>
      );

    default:
      return <HiOutlineDocumentText className={className} />;
  }
};

export default OperationIcon;
