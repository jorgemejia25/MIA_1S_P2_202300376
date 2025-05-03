interface CommandResponse {
  output: string;
}

export const sendCommand = async (command: string): Promise<string> => {
  console.log(command);

  const apiUrl = process.env.API_URL || "http://3.85.93.122:8080";

  const response = await fetch(`${apiUrl}/command`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ command }),
  });

  const data: CommandResponse = await response.json();
  return data.output;
};
