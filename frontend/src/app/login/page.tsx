"use client";

import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import Link from "next/link";
import { Fieldset, Field, ErrorMessage } from "@/components/ui/fieldset";
import { useRequest } from "@/hooks/useRequest";
import { useRouter } from "next/navigation";
import toast from "react-hot-toast";
import { useUserStore } from "@/store/useUserStore";

const formSchema = z.object({
  identifier: z.string().min(1, "Email or username is required"),
  password: z.string().min(8, "Password must be at least 8 characters"),
  rememberMe: z.boolean().optional(),
});

type FormValues = z.infer<typeof formSchema>;

export default function LoginPage() {
  const {
    register,
    handleSubmit,
    formState: { errors },
    watch,
    setValue,
  } = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      identifier: "",
      password: "",
      rememberMe: false,
    },
  });

  const { isLoading, data, error, post } = useRequest();
  const router = useRouter();
  const login = useUserStore((state) => state.login);

  // Handle store hydration
  useEffect(() => {
    useUserStore.persist.rehydrate();
  }, []);

  // Handle login errors - make toast persist longer
  useEffect(() => {
    if (error) {
      // Log the error object to inspect its structure
      console.log("Login error object:", error);

      // Check for specific HTTP status codes in the error message
      let errorMessage = "Sign in failed. Please try again.";

      let httpStatus: number | undefined = undefined;

      // Attempt to get status from error.response.status (common for axios)
      if (
        typeof error === "object" &&
        error !== null &&
        (error as any).response &&
        typeof (error as any).response.status === "number"
      ) {
        httpStatus = (error as any).response.status;
      }
      // ELSE, if useRequest puts status directly on error object
      else if (
        typeof error === "object" &&
        error !== null &&
        typeof (error as any).status === "number"
      ) {
        httpStatus = (error as any).status;
      }
      // Add other checks here based on what console.log(error) reveals
      // For example, if error is just a string:
      // else if (typeof error === 'string' && error.includes('401')) {
      //   httpStatus = 401;
      // }

      console.log("Extracted HTTP status:", httpStatus); // See what status was extracted

      if (httpStatus === 401) {
        errorMessage = "Invalid email/username or password.";
      } else if (httpStatus === 400) {
        errorMessage = "Please check your input and try again.";
      } else if (httpStatus === 500) {
        errorMessage = "Server error. Please try again later.";
      } else if (
        error.message &&
        typeof error.message === "string" &&
        (error.message.includes("Session expired") ||
          error.message.includes("Invalid credentials")) // Key check
      ) {
        // If httpStatus wasn't found, check the error message string.
        errorMessage = "Invalid email/username or password.";
      }

      // Use a longer duration and prevent duplicate toasts
      toast.error(errorMessage, {
        duration: 5000, // 5 seconds
        id: "login-error", // Prevent duplicate toasts
      });
    }
  }, [error]);

  return (
    <div className="max-w-md mx-auto my-12 p-6 bg-white rounded-lg shadow-md dark:bg-zinc-900">
      <h1 className="text-2xl font-bold mb-6 text-center">Sign In</h1>
      <form
        onSubmit={handleSubmit((data: FormValues) => {
          post("/api/auth/signin", data, (data) => {
            login(data.user);
            toast.success("Logged in successfully!");
            console.log("User logged in:", { data });
            router.push("/profile");
          });
        })}
      >
        <Fieldset className="space-y-6">
          <Field>
            <label
              className="block text-sm font-medium mb-1"
              htmlFor="identifier"
            >
              Email or Username *
            </label>
            <Input
              id="identifier"
              type="text"
              placeholder="Email or username"
              {...register("identifier")}
            />
            {errors.identifier?.message && (
              <ErrorMessage>{errors.identifier.message}</ErrorMessage>
            )}
          </Field>

          <Field>
            <label
              className="block text-sm font-medium mb-1"
              htmlFor="password"
            >
              Password *
            </label>
            <Input
              id="password"
              type="password"
              placeholder="••••••••"
              {...register("password")}
            />
            {errors.password?.message && (
              <ErrorMessage>{errors.password.message}</ErrorMessage>
            )}
          </Field>

          <Field>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="rememberMe"
                  {...register("rememberMe")}
                  className="hidden"
                />
                <div
                  onClick={() => setValue("rememberMe", !watch("rememberMe"))}
                  className={`w-5 h-5 rounded border cursor-pointer ${
                    watch("rememberMe")
                      ? "bg-blue-600 border-blue-600"
                      : "border-gray-300 dark:border-gray-600"
                  }`}
                >
                  {watch("rememberMe") && (
                    <svg
                      className="w-4 h-4 text-white mx-auto"
                      viewBox="0 0 20 20"
                      fill="currentColor"
                    >
                      <path
                        fillRule="evenodd"
                        d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                        clipRule="evenodd"
                      />
                    </svg>
                  )}
                </div>
                <label htmlFor="rememberMe" className="text-sm">
                  Remember me
                </label>
              </div>
              <Link
                href="/forgot-password"
                className="text-sm text-blue-600 hover:underline"
              >
                Forgot password?
              </Link>
            </div>
          </Field>
        </Fieldset>

        <div className="mt-6">
          <Button
            type="submit"
            className="w-full bg-blue-600 hover:bg-blue-700"
            disabled={isLoading}
          >
            {isLoading ? "Signing in..." : "Sign In"}
          </Button>
        </div>
      </form>

      <p className="mt-4 text-center text-sm">
        Don't have an account?{" "}
        <Link href="/register" className="text-blue-600 hover:underline">
          Register
        </Link>
      </p>
    </div>
  );
}
