interface CommandResponse {
  output: string;
}

export const sendCommand = async (command: string): Promise<string> => {
  console.log(command);

  const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://54.196.151.70:8080";

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
