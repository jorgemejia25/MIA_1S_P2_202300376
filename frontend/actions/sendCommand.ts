"use server";

interface CommandResponse {
  output: string;
}

export const sendCommand = async (command: string): Promise<string> => {
  console.log(command);

  const response = await fetch("http://localhost:8080/command", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ command }),
  });

  const data: CommandResponse = await response.json();
  return data.output;
};
