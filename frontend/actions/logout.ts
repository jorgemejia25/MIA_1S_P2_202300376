interface LogoutResponse {
  success?: boolean;
  msg?: string;
}

export const logout = async (): Promise<LogoutResponse> => {
  try {
    const apiUrl = process.env.API_URL || "http://3.85.93.122:8080";
    const response = await fetch(`${apiUrl}/logout`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
    });

    const data: LogoutResponse = await response.json();
    return data;
  } catch (error) {
    console.error("Logout error:", error);
    return {
      success: false,
      msg: "Error de conexión con el servidor",
    };
  }
};
