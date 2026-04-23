import type { JSX } from "react";
import {
  isRouteErrorResponse,
  Links,
  Meta,
  Outlet,
  Scripts,
  ScrollRestoration,
} from "react-router";
import { ApolloClient, HttpLink, InMemoryCache } from "@apollo/client";
import { ApolloProvider } from "@apollo/client/react";
import {
  Box,
  Code,
  Container,
  ColorSchemeScript,
  Flex,
  MantineProvider,
  Text,
  Title,
  mantineHtmlProps,
} from "@mantine/core";
import type { Route } from "./+types/root";
import { theme } from "./theme";
import "./css/app.css";

const client = new ApolloClient({
  link: new HttpLink({ uri: "/query" }),
  cache: new InMemoryCache(),
});

export const links: Route.LinksFunction = () => [
  // Fonts
  {
    rel: "preconnect",
    href: "https://fonts.googleapis.com",
  },
  {
    rel: "preconnect",
    href: "https://fonts.gstatic.com",
    crossOrigin: "anonymous",
  },
  {
    rel: "stylesheet",
    href: "https://fonts.googleapis.com/css2?family=Open+Sans:ital,wght@0,300..800;1,300..800&family=Yellowtail&display=swap",
  },
];

export function Layout({
  children,
}: {
  children: React.ReactNode;
}): JSX.Element {
  return (
    <html lang="en" {...mantineHtmlProps}>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <ColorSchemeScript />
        <Meta />
        <Links />
        <title>Environmental Monitor</title>
      </head>
      <body>
        <ApolloProvider client={client}>
          <MantineProvider theme={theme}>{children}</MantineProvider>
        </ApolloProvider>
        <ScrollRestoration />
        <Scripts />
      </body>
    </html>
  );
}

export default function App(): JSX.Element {
  return (
    <Container pt="md" pb="lg">
      <Title order={1} ta="center" mb="md">
        Environmental Monitor
      </Title>

      <Flex direction="column" gap="lg">
        <Outlet />
      </Flex>
    </Container>
  );
}

export function HydrateFallback(): JSX.Element {
  return <></>;
}

export function ErrorBoundary({
  error,
}: Route.ErrorBoundaryProps): JSX.Element {
  let message = "Oops!";
  let details = "An unexpected error occurred.";
  let stack: string | undefined;

  if (isRouteErrorResponse(error)) {
    message = error.status === 404 ? "404" : "Error";
    details =
      error.status === 404
        ? "The requested page could not be found."
        : error.statusText || details;
  } else if (import.meta.env.DEV && error && error instanceof Error) {
    details = error.message;
    stack = error.stack;
  }

  return (
    <Container component="main" pt="xl" p="md" mx="auto">
      <Title>{message}</Title>
      <Text>{details}</Text>
      {stack && (
        <Box component="pre" w="100%" style={{ overflowX: "auto" }} p="md">
          <Code>{stack}</Code>
        </Box>
      )}
    </Container>
  );
}
