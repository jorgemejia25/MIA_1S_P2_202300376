import React from "react";

interface ErrorMessageProps {
  message: string | null;
}

const ErrorMessage: React.FC<ErrorMessageProps> = ({ message }) => {
  if (!message) return null;

  return (
    <div className="mb-6 p-4 bg-red-900/30 border border-red-700 rounded-md text-red-300">
      <p className="font-medium">Error: {message}</p>
    </div>
  );
};

export default ErrorMessage;
