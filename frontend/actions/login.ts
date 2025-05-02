"use server";

interface LoginResponse {
  success?: boolean;
  msg?: string;
}

export const login = async (
  prevState: LoginResponse,
  formData: FormData
): Promise<LoginResponse> => {
  const partition = formData.get("partition") as string;
  const username = formData.get("username") as string;
  const password = formData.get("password") as string;

  console.log(`Login attempt: ${username} on partition ${partition}`);

  try {
    const response = await fetch("http://localhost:8080/login", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ partition, username, password }),
    });

    const data: LoginResponse = await response.json();

    // Si la autenticación fue exitosa, podríamos establecer una cookie o estado aquí
    if (data.success) {
      console.log("Login successful");
    }

    return data;
  } catch (error) {
    console.error("Login error:", error);
    return {
      success: false,
      msg: "Error de conexión con el servidor",
    };
  }
};
