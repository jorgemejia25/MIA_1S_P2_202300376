"use server";

interface LogoutResponse {
  success?: boolean;
  msg?: string;
}

export const logout = async (): Promise<LogoutResponse> => {
  console.log("Logout attempt");

  try {
    const response = await fetch("http://localhost:8080/logout", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
    });

    const data: LogoutResponse = await response.json();

    if (data.success) {
      console.log("Logout successful");
    }

    return data;
  } catch (error) {
    console.error("Logout error:", error);
    return {
      success: false,
      msg: "Error de conexión con el servidor",
    };
  }
};
